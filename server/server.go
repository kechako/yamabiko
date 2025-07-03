package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kechako/seezlog"
	"github.com/kechako/skkdic"
	"github.com/kechako/yamabiko/config"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	f   *os.File
	l   *slog.Logger
	cfg *config.Config
	d   *skkdic.Dictionary

	mu         sync.Mutex
	activeConn map[*net.Conn]struct{}
	wg         sync.WaitGroup
}

func New(cfg *config.Config) (*Server, error) {
	s := &Server{
		d:   skkdic.New(),
		cfg: cfg,
	}

	if err := s.init(cfg); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) init(cfg *config.Config) error {
	if err := s.buildLogger(cfg); err != nil {
		return err
	}

	for _, dict := range cfg.Dictionaries {
		err := s.d.ReadFile(dict.Path, skkdic.Add, skkdic.WithInputEncoding(dict.Encoding))
		if err != nil {
			return fmt.Errorf("failed to read dictionary file %s: %w", dict.Path, err)
		}
	}

	return nil
}

func (s *Server) Close() error {
	var errs []error

	if s.f != nil {
		if err := s.f.Close(); err != nil {
			errs = append(errs, err)
		}
		s.f = nil
	}

	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return errors.Join(errs...)
	}

	return nil
}

func (s *Server) Serve(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	addr := makeAddr(s.cfg)

	var c net.ListenConfig
	l, err := c.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	defer l.Close()

	s.l.Info("Server started",
		slog.String("address", addr),
		slog.String("send_encoding", string(s.cfg.SendEncoding)),
		slog.String("recv_encoding", string(s.cfg.RecvEncoding)))

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		<-ctx.Done()

		l.Close()

		for conn := range s.activeConn {
			(*conn).Close()
			s.setActiveConn(conn, false)
		}

		return nil
	})

	g.Go(func() error {
		var tempDelay time.Duration
	loop:
		for {
			c, err := l.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					break loop
				default:
				}
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					if tempDelay == 0 {
						tempDelay = 5 * time.Millisecond
					} else {
						tempDelay *= 2
					}
					if max := 1 * time.Second; tempDelay > max {
						tempDelay = max
					}
					time.Sleep(tempDelay)
					continue
				}
				return fmt.Errorf("failed to accept connection: %w", err)
			}
			tempDelay = 0
			s.setActiveConn(&c, true)
			s.wg.Add(1)
			go s.serve(ctx, c)
		}

		s.wg.Wait()

		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

const (
	ClientEnd        = '0'
	ClientRequest    = '1'
	ClientVersion    = '2'
	ClientHost       = '3'
	ClientCompletion = '4'

	ServerError    = '0'
	ServerFound    = '1'
	ServerNotFound = '4'
)

func (s *Server) serve(ctx context.Context, conn net.Conn) {
	defer s.wg.Done()
	defer s.setActiveConn(&conn, false)
	defer conn.Close()

	logger := s.l.With(slog.String("remote", conn.RemoteAddr().String()))
	logger.Info("New connection established")
	defer logger.Info("Connection closed")

	sendEnc := s.cfg.SendEncoding
	recvEnc := s.cfg.RecvEncoding
	w := sendEnc.NewEncoder().Writer(conn)
	r := recvEnc.NewDecoder().Reader(conn)

	buf := make([]byte, 1024)
	var ret bytes.Buffer
	ret.Grow(4096)
loop:
	for {
		ret.Reset()

		n, err := r.Read(buf)
		if err != nil {
			select {
			case <-ctx.Done():
				break loop
			default:
			}
			if err == io.EOF {
				break loop
			}
			logger.Error("failed to read command", slog.Any("error", err))
			return
		}
		cmd := string(buf[:n])
		switch cmd[0] {
		case ClientEnd:
			logger.Debug("CLIENT END")
			break loop
		case ClientRequest:
			i := strings.IndexByte(cmd, ' ')
			if i < 0 {
				i = strings.IndexByte(cmd, '\n')
			}
			if i < 0 {
				i = len(cmd)
			}

			key := cmd[1:i]
			logger.Debug("REQUEST", slog.String("key", key))

			candidates := s.d.Lookup(key)
			if isEmpty(candidates) {
				ret.WriteRune(ServerNotFound)
				ret.WriteString(cmd[1:])
				logger.Debug("REQUEST: not found", slog.String("key", key))
			} else {
				ret.WriteRune(ServerFound)
				for c := range s.d.Lookup(key) {
					ret.WriteRune('/')
					ret.WriteString(c.String())
				}
				ret.WriteString("/\n")
				if logger.Enabled(ctx, slog.LevelDebug) {
					logger.Debug("REQUEST", slog.String("candidate", ret.String()))
				}
			}
		case ClientVersion:
			logger.Debug("VERSION")
			ret.WriteString(Version)
		case ClientHost:
			logger.Debug("HOST")
			ret.WriteString(conn.LocalAddr().String())
		case ClientCompletion:
			logger.Debug("COMPLETION")
			ret.WriteRune(ServerFound)
			ret.WriteString("//\n")
		default:
			logger.Warn("UNKNOWN COMMAND", slog.String("command", cmd))
			continue
		}
		if _, err := w.Write(ret.Bytes()); err != nil {
			logger.Error("failed to write response", slog.String("response", ret.String()), slog.Any("error", err))
		}
	}
}

func (s *Server) setActiveConn(conn *net.Conn, set bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeConn == nil {
		s.activeConn = make(map[*net.Conn]struct{})
	}

	if set {
		s.activeConn[conn] = struct{}{}
	} else {
		delete(s.activeConn, conn)
	}
}

func makeAddr(cfg *config.Config) string {
	return net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
}

func isEmpty[T any](sec iter.Seq[T]) bool {
	for range sec {
		return false
	}
	return true
}

func (s *Server) buildLogger(cfg *config.Config) error {
	var logWriter io.Writer
	if cfg.Logging.Path == "" {
		logWriter = os.Stdout
	} else {
		file, err := os.Create(cfg.Logging.Path)
		if err != nil {
			return fmt.Errorf("failed to create log file: %w", err)
		}
		s.f = file
		logWriter = file
	}

	var handler slog.Handler
	if cfg.Logging.JSON {
		handler = slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
			Level: cfg.Logging.Level,
		})
	} else {
		handler = seezlog.NewHandler(logWriter, cfg.Logging.Path == "", &slog.HandlerOptions{
			Level: cfg.Logging.Level,
		})
	}

	s.l = slog.New(handler)

	return nil
}
