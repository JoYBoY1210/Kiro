package main

import (
	"fmt"
	// "io"
	// "io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/JoYBoY1210/kiro/proxy"
	"github.com/JoYBoY1210/kiro/services"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Services []Service `yaml:"services"`
	Mesh     Mesh      `yaml:"mesh"`
}

type Service struct {
	Name      string `yaml:"name"`
	Port      int    `yaml:"port"`
	ProxyPort int    `yaml:"proxyPort"`
}

type Mesh struct {
	MTLS         string         `yaml:"mTLS"`
	Logging      string         `yaml:"logging"`
	AllowedCalls []AllowedCalls `yaml:"allowedCalls"`
	Auth         Auth           `yaml:"auth"`
	Headers      Headers        `yaml:"headers"`
}
type Auth struct {
	Mode                string `yaml:"mode"`
	HMACSecret          string `yaml:"hmacSecret"`
	ClockSkew           int    `yaml:"clockSkew"`
	RequiredLocalCaller bool   `yaml:"requiredLocalCaller"`
}
type Headers struct {
	Identity  string `yaml:"identity"`
	Caller    string `yaml:"caller"`
	Timestamp string `yaml:"timestamp"`
	Signature string `yaml:"signature"`
}

type AllowedCalls struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

func buildServiceMap(servicesCfg []Service) map[string]int {
	serviceMap := make(map[string]int)
	for _, s := range servicesCfg {
		serviceMap[strings.ToLower(s.Name)] = s.ProxyPort
	}
	return serviceMap

}

func buildAllowed(meshCfg Mesh) map[string]map[string]bool {
	m := make(map[string]map[string]bool)
	for _, ac := range meshCfg.AllowedCalls {
		if _, ok := m[strings.ToLower(ac.From)]; !ok {
			m[strings.ToLower(ac.From)] = make(map[string]bool)
		}
		m[strings.ToLower(ac.From)][strings.ToLower(ac.To)] = true
	}
	return m
}

func loadConfig(path string) (Config, error) {
	var config Config
	file, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return config, err
	}
	fmt.Println("config loaded")
	return config, nil
}

func main() {
	config, err := loadConfig("config/config.yaml")
	if err != nil {
		log.Println("config could not be loaded, error: ", err)
		return
	}
	serviceMap := buildServiceMap(config.Services)
	allowedCallsMap := buildAllowed(config.Mesh)
	// fmt.Println("Allowed map contents:")
	// for from, tos := range allowedCallsMap {
	// 	for to := range tos {
	// 		fmt.Printf("  %s -> %s\n", from, to)
	// 	}
	// }

	fmt.Println("services: ")
	for _, s := range config.Services {
		fmt.Printf("- %s, port: %d, proxyPort: %d", s.Name, s.Port, s.ProxyPort)
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
	fmt.Println("Mesh config: ")
	fmt.Println("mTLS: ", config.Mesh.MTLS)
	fmt.Println("logging: ", config.Mesh.Logging)
	fmt.Println("allowedCalls: ")
	for _, ac := range config.Mesh.AllowedCalls {
		fmt.Printf("- from: %s, to: %s", ac.From, ac.To)
		fmt.Printf("\n")

	}

	var wg sync.WaitGroup
	for _, s := range config.Services {
		wg.Add(1)
		go func(s Service) {
			defer wg.Done()
			if s.Name == "authService" {
				services.StartAuthService(s.Port, s.Name, s.ProxyPort, "./security/certs/auth.crt", "./security/certs/auth.key", "./security/certs/ca.crt")
			} else if s.Name == "dashboardService" {
				services.StartDashboardService(s.Port, s.Name, s.ProxyPort, "./security/certs/dashboard.crt", "./security/certs/dashboard.key", "./security/certs/ca.crt")
			} else if s.Name == "profileService" {
				services.StartProfileService(s.Port, s.Name, s.ProxyPort, "./security/certs/profile.crt", "./security/certs/profile.key", "./security/certs/ca.crt")
			} else {
				fmt.Println("unknown service")
			}

		}(s)
		wg.Add(1)
		go func(s Service) {
			defer wg.Done()
			p := proxy.Proxy{
				ServiceName:         strings.ToLower(s.Name),
				ListenPort:          s.ProxyPort,
				TargetPort:          s.Port,
				ServiceMap:          serviceMap,
				Allowed:             allowedCallsMap,
				AuthMode:            config.Mesh.Auth.Mode,
				HMACSecret:          []byte(config.Mesh.Auth.HMACSecret),
				ClockSkew:           config.Mesh.Auth.ClockSkew,
				RequiredLocalCaller: config.Mesh.Auth.RequiredLocalCaller,
				HeaderIdentity:      config.Mesh.Headers.Identity,
				HeaderTimestamp:     config.Mesh.Headers.Timestamp,
				HeaderSignature:     config.Mesh.Headers.Signature,
				HeaderCallerChain:   config.Mesh.Headers.Caller,
				CertFile:            "./security/certs/" + s.Name + ".crt",
				KeyFile:             "./security/certs/" + s.Name + ".key",
				CAFile:              "./security/certs/ca.crt",
			}

			p.Start()

		}(s)
	}
	// client := services.MeshClient(5001, "dashboardService", "./security/certs/dashboard.crt", "./security/certs/dashboard.key", "./security/certs/ca.crt")
	// resp, err := client.Get("http://authService/auth")
	// if err != nil {
	// 	fmt.Println("err: ", err)
	// }
	// defer resp.Body.Close()
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Println("err: ", err)
	// }
	// fmt.Println("body: ", string(body))

	wg.Wait()
}
