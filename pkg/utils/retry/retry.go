// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package retry

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"
)

// const constants
var (
	ErrorAbort                 = errors.New("stop retry")
	ErrorTimeout               = errors.New("retry timeout")
	ErrorContextDeadlineExceed = errors.New("context deadline exceeded")
	ErrorEmptyRetryFunc        = errors.New("empty retry function")
	ErrorTimeFormat            = errors.New("time out err")
)

// define a retry func
type RetriesFunc func() error

// option func
type Option func(c *Config)

// define hook func
type HookFunc func()

// define a retry check func
type RetriesChecker func(err error) (needRetry bool)

// type a config struct
type Config struct {
	MaxRetryTimes int
	Timeout       time.Duration
	RetryChecker  RetriesChecker
	Strategy      Strategy
	RecoverPanic  bool
	BeforeTry     HookFunc
	AfterTry      HookFunc
}

// const constants
var (
	DefaultMaxRetryTimes = 3
	DefaultTimeout       = time.Minute
	DefaultInterval      = time.Second * 2
	DefaultRetryChecker  = func(err error) bool {
		return !errors.Is(err, ErrorAbort) // not abort error, should continue retry
	}
)

// create a default config
func newDefaultConfig() *Config {
	return &Config{
		MaxRetryTimes: DefaultMaxRetryTimes,
		RetryChecker:  DefaultRetryChecker,
		Timeout:       DefaultTimeout,
		Strategy:      NewLinear(DefaultInterval),
		BeforeTry:     func() {},
		AfterTry:      func() {},
	}
}

// with time out
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// with max retry times
func WithMaxRetryTimes(times int) Option {
	return func(c *Config) {
		c.MaxRetryTimes = times
	}
}

// with recover panic
func WithRecoverPanic() Option {
	return func(c *Config) {
		c.RecoverPanic = true
	}
}

// a before hook
func WithBeforeHook(hook HookFunc) Option {
	return func(c *Config) {
		c.BeforeTry = hook
	}
}

// a hook with after
func WithAfterHook(hook HookFunc) Option {
	return func(c *Config) {
		c.AfterTry = hook
	}
}

// retry check
func WithRetryChecker(checker RetriesChecker) Option {
	return func(c *Config) {
		c.RetryChecker = checker
	}
}

// with backoff strategy
func WithBackOffStrategy(s BackoffStrategy, duration time.Duration) Option {
	return func(c *Config) {
		switch s {
		case StrategyConstant:
			c.Strategy = NewConstant(duration)
		case StrategyLinear:
			c.Strategy = NewLinear(duration)
		case StrategyFibonacci:
			c.Strategy = NewFibonacci(duration)
		}
	}
}

// with custom strategy
func WithCustomStrategy(s Strategy) Option {
	return func(c *Config) {
		c.Strategy = s
	}
}

// do
func Do(ctx context.Context, fn RetriesFunc, opts ...Option) error {
	if fn == nil {
		return ErrorEmptyRetryFunc
	}
	var (
		abort         = make(chan struct{}, 1) // caller choose to abort retry
		run           = make(chan error, 1)
		panicInfoChan = make(chan string, 1)

		timer  *time.Timer
		runErr error
	)
	config := newDefaultConfig()
	for _, o := range opts {
		o(config)
	}
	if config.Timeout > 0 {
		timer = time.NewTimer(config.Timeout)
	} else {
		return ErrorTimeFormat
	}
	go func() {
		var err error
		defer func() {
			if e := recover(); e == nil {
				return
			} else {
				panicInfoChan <- fmt.Sprintf("retry function panic has occurred, err=%v, stack:%s", e, string(debug.Stack()))
			}
		}()
		for i := 0; i < config.MaxRetryTimes; i++ {
			config.BeforeTry()
			err = fn()
			config.AfterTry()
			if err == nil {
				run <- nil
				return
			}
			// check whether to retry
			if config.RetryChecker != nil {
				needRetry := config.RetryChecker(err)
				if !needRetry {
					abort <- struct{}{}
					return
				}
			}
			if config.Strategy != nil {
				interval := config.Strategy.Sleep(i + 1)
				<-time.After(interval)
			}
		}
		run <- err
	}()
	select {
	case <-ctx.Done():
		// context deadline exceed
		return ErrorContextDeadlineExceed
	case <-timer.C:
		// timeout
		return ErrorTimeout
	case <-abort:
		// caller abort
		return ErrorAbort
	case msg := <-panicInfoChan:
		// panic occurred
		if !config.RecoverPanic {
			panic(msg)
		}
		runErr = fmt.Errorf("panic occurred=%s", msg)
	case e := <-run:
		// normal run
		if e != nil {
			runErr = fmt.Errorf("retry failed, err=%w", e)
		}
	}
	return runErr
}
