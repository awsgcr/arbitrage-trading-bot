package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	windows = "windows"
	linux   = "linux"
)

var (
	goarch  string
	goos    string
	gocc    string
	cgo     bool
	pkgArch string
	race    bool
	isDev   = false
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	ensureGoPath()

	flag.StringVar(&goarch, "goarch", runtime.GOARCH, "GOARCH")
	flag.StringVar(&goos, "goos", runtime.GOOS, "GOOS")
	flag.StringVar(&gocc, "cc", "", "CC")
	flag.BoolVar(&cgo, "cgo-enabled", cgo, "Enable cgo")
	flag.StringVar(&pkgArch, "pkg-arch", "", "PKG ARCH")
	flag.BoolVar(&race, "race", race, "Use race detector")
	flag.BoolVar(&isDev, "dev", isDev, "optimal for development, skips certain steps")
	flag.Parse()

	if pkgArch == "" {
		pkgArch = goarch
	}

	if flag.NArg() == 0 {
		log.Println("Usage: go run build.go build")
		return
	}

	log.Println("isDev:", isDev)

	for _, cmd := range flag.Args() {
		switch cmd {
		case "setup":
			setup()

		case "coin_labor":
			clean()
			build("coin_labor", "./pkg/cmd/coin_labor", []string{})

		case "clean":
			clean()

		default:
			log.Fatalf("Unknown command %q", cmd)
		}
	}
}

func ensureGoPath() {
	if os.Getenv("GOPATH") == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		gopath := filepath.Clean(filepath.Join(cwd, "../../../../"))
		log.Println("GOPATH is", gopath)
		os.Setenv("GOPATH", gopath)
	}
}

func setup() {
	//runPrint("go", "get", "-v", "github.com/golang/dep")
	runPrint("go", "install", "-v", "./pkg/cmd/cli")
}

func build(binaryName, pkg string, tags []string) {
	binary := fmt.Sprintf("./bin/%s-%s/%s", goos, goarch, binaryName)
	if isDev {
		//don't include os and arch in output path in dev environment
		binary = fmt.Sprintf("./bin/%s", binaryName)
	}

	if goos == windows {
		binary += ".exe"
	}

	if !isDev {
		rmr(binary, binary+".md5")
	}
	//args := []string{"build", "-ldflags", ldflags()}
	args := []string{"build"}
	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}
	if race {
		args = append(args, "-race")
	}

	args = append(args, "-o", binary)
	args = append(args, pkg)

	if !isDev {
		setBuildEnv()
		runPrint("go", "version")
		fmt.Printf("Targeting %s/%s\n", goos, goarch)
	}

	runPrint("go", args...)

	if !isDev {
		// Create an md5 checksum of the binary, to be included in the archive for automatic upgrades.
		err := md5File(binary)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func rmr(paths ...string) {
	for _, path := range paths {
		log.Println("rm -r", path)
		os.RemoveAll(path)
	}
}

func clean() {
	if isDev {
		return
	}

	rmr(filepath.Join(os.Getenv("GOPATH"), fmt.Sprintf("pkg/%s_%s/jasonzhu.com/coin_labor", goos, goarch)))
}

func setBuildEnv() {
	_ = os.Setenv("GOOS", goos)
	if goos == windows {
		// require windows >=7
		_ = os.Setenv("CGO_CFLAGS", "-D_WIN32_WINNT=0x0601")
	}
	if goarch != "amd64" || goos != linux {
		// needed for all other archs
		cgo = true
	}
	if strings.HasPrefix(goarch, "armv") {
		_ = os.Setenv("GOARCH", "arm")
		_ = os.Setenv("GOARM", goarch[4:])
	} else {
		_ = os.Setenv("GOARCH", goarch)
	}
	if goarch == "386" {
		_ = os.Setenv("GO386", "387")
	}
	if cgo {
		_ = os.Setenv("CGO_ENABLED", "1")
	}
	if gocc != "" {
		_ = os.Setenv("CC", gocc)
	}
}

func runPrint(cmd string, args ...string) {
	log.Println(cmd, strings.Join(args, " "))
	ecmd := exec.Command(cmd, args...)
	ecmd.Stdout = os.Stdout
	ecmd.Stderr = os.Stderr
	err := ecmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func md5File(file string) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()

	h := md5.New()
	_, err = io.Copy(h, fd)
	if err != nil {
		return err
	}

	out, err := os.Create(file + ".md5")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "%x\n", h.Sum(nil))
	if err != nil {
		return err
	}

	return out.Close()
}
