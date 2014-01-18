/* Goiardi configuration. */

/*
 * Copyright (c) 2013-2014, Jeremy Bingham (<jbingham@gmail.com>)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"github.com/jessevdk/go-flags"
	"github.com/BurntSushi/toml"
	"os"
	"log"
	"fmt"
)

/* Master struct for configuration. */
type Conf struct {
	Ipaddress string
	Port int
	Hostname string
	ConfFile string
	IndexFile string
	DataStoreFile string
	DebugLevel int
	FreezeInterval int
	FreezeData bool
	LogFile string
}

/* Struct for command line options. */
type Options struct {
	Version bool `short:"v" long:"version" description:"Print version info."`
	Verbose []bool `short:"V" long:"verbose" description:"Show verbose debug information. (not implemented)"`
	ConfFile string `short:"c" long:"config" description:"Specify a config file to use."`
	Ipaddress string `short:"I" long:"ipaddress" description:"Listen on a specific IP address."`
	Hostname string `short:"H" long:"hostname" description:"Hostname to use for this server. Defaults to hostname reported by the kernel."`
	Port int `short:"P" long:"port" description:"Port to listen on. (default: 4545)"`
	IndexFile string `short:"i" long:"index-file" description:"File to save search index data to."`
	DataStoreFile string `short:"D" long:"data-file" description:"File to save data store data to."`
	FreezeInterval int `short:"F" long:"freeze-interval" description:"Interval in seconds to freeze in-memory data structures to disk (requires -i/--index-file and -D/--data-file options to be set). (Default 300 seconds/5 minutes.)"`
	LogFile string `short:"L" long:"log-file" description:"Log to file X"`
}

// The goiardi version
const Version = "0.3.0"
// The chef version we're at least aiming for, even if it's not complete yet
const ChefVersion = "11.0.8"

/* The general plan is to read the command-line options, then parse the config
 * file, fill in the config struct with those values, then apply the 
 * command-line options to the config struct. We read the cli options first so
 * we know to look for a different config file if needed, but otherwise the
 * command line options override what's in the config file. */

func InitConfig() *Conf { return &Conf{ } }

var Config = InitConfig()

// Read and apply arguments from the command line.
func ParseConfigOptions() error {
	var opts = &Options{ }
	_, err := flags.Parse(opts)

	if err != nil {
		if err.(*flags.Error).Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			log.Println(err)
			os.Exit(1)
		}
	}

	if opts.Version {
		fmt.Printf("goiardi version %s (aiming for compatibility with Chef Server version %s).\n", Version, ChefVersion)
		os.Exit(0)
	}

	/* Load the config file. Command-line options have precedence over
	 * config file options. */
	if opts.ConfFile != "" {
		if _, err := toml.DecodeFile(opts.ConfFile, Config); err != nil {
			panic(err)
			os.Exit(1)
		}
		Config.FreezeData = false
	}
	
	if opts.Hostname != "" {
		Config.Hostname = opts.Hostname
	} else {
		if Config.Hostname == "" {
			Config.Hostname, err = os.Hostname()
			if err != nil {
				log.Println(err)
				Config.Hostname = "localhost"
			}
		}
	}

	if !((opts.DataStoreFile == "" && opts.IndexFile == "") || (opts.DataStoreFile != "" && opts.IndexFile != "")) {
		err := fmt.Errorf("-i and -D must either both be specified, or not specified.")
		panic(err)
		os.Exit(1)
	}
	if opts.DataStoreFile != "" {
		Config.DataStoreFile = opts.DataStoreFile
		Config.FreezeData = true
	}

	if opts.IndexFile != "" {
		Config.IndexFile = opts.IndexFile
		Config.FreezeData = true
	}

	if opts.LogFile != "" {
		Config.LogFile = opts.LogFile
	}
	if Config.LogFile != "" {
		lfp, lerr := os.Create(Config.LogFile)
		if lerr != nil {
			log.Println(err)
			os.Exit(1)
		}
		log.SetOutput(lfp)
	}

	if !Config.FreezeData && (opts.FreezeInterval != 0 || Config.FreezeInterval != 0) {
		log.Printf("FYI, setting the freeze data interval's not especially useful without setting the index and data files.")
	}
	if opts.FreezeInterval != 0 {
		Config.FreezeInterval = opts.FreezeInterval
	}
	if Config.FreezeInterval == 0 {
		Config.FreezeInterval = 300
	}

	Config.Ipaddress = opts.Ipaddress
	if opts.Port != 0 {
		Config.Port = opts.Port
	}
	if Config.Port == 0 {
		Config.Port = 4545
	}
	Config.DebugLevel = len(opts.Verbose)

	return nil
}

// The address and port goiardi is configured to listen on.
func ListenAddr() string {
	listen_addr := fmt.Sprintf("%s:%d", Config.Ipaddress, Config.Port)
	return listen_addr
}

// The hostname and port goiardi is configured to use.
func ServerHostname() string {
	hostname := fmt.Sprintf("%s:%d", Config.Hostname, Config.Port)
	return hostname
}

// The base URL
func ServerBaseURL() string {
	/* TODO: allow configuring using http vs. https */
	url := fmt.Sprintf("http://%s", ServerHostname())
	return url
}