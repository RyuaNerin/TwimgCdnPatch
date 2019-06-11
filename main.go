package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/mkideal/cli"
)

const (
	twimgHostName 	= "pbs.twimg.com";
	twimgTestURI	= "https://pbs.twimg.com/media/CgAc2lSUMAA30oE.jpg:orig"

	pingCount		= 20
	pingTimeout		= 1000
	httpBufferSize	= 32 * 1024
	httpCount		= 50
	httpTimeout		= 10 * 1000
)

var (
	regPattern = regexp.MustCompile(fmt.Sprintf("^[^\\s\\t#].+[\\s\\t]+%s.*$", strings.ReplaceAll(twimgHostName, ".", "\\.")))
)

type args struct {
	Help		bool	`cli:"h,help"      usage:"도움말을 표시합니다"`
	Install		bool	`cli:"i,install"   usage:"twimg dns 패치를 합니다"`
	Uninstall	bool	`cli:"u,uninstall" usage:"twimg dns 패치를 제거합니다"`
}
func (args *args) AutoHelp() bool {
	return args.Help
}

func main() {
	os.Exit(cli.Run(new(args), func(ctx *cli.Context) error {
		appmain(ctx.Argv().(*args))
		return nil
	}))
}

func appmain(args *args) {
    defer func() {
        if err := recover(); err != nil {
			switch err.(type) {
			case error:
				//println(err.(error).Error())
				panic(err)
			case string:
				println(err.(string))
			}
        }
	}()

	patch(args)
}

func patch(args *args) {
	var bestCdnAddr string

	fs, err := os.OpenFile(getHostsPath(), os.O_RDWR, 0)
	if err != nil {
		panic(err)
	}
	defer func() {
		fs.Sync()
		fs.Close()
	}()

	if !args.Install || !args.Uninstall {
		println("twimg 패치를 확인중입니다")
	}
	
	lines, patched := readAllHosts(fs)

	if args.Install && patched {
		println("이미 Twimg dns 패치가 되어있습니다.")
		return
	}
	if args.Uninstall && !patched {
		println("Twimg dns 패치가 되어있지 않습니다.")
		return
	}

	if patched {
		// WriteLine 이라 어처피 한 줄 더 추가됨
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines) - 1]) == "" {
			lines = lines[:len(lines) - 1]
		}
	} else {
		println("Twimg cdn 서버 정보를 가져오고 있습니다.")
		hosts := getAddresses()
		
		println("가장 연결 상태가 좋은 CDN 을 찾고 있습니다")
		bestCdnAddr = getBestCdn(hosts)

		if bestCdnAddr == "" {
			panic("CDN 테스트를 실패하였습니다.\n나중에 다시 시도해주세요.")
		}
		
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines) - 1]) != "" {
			lines = append(lines, "")
		}
		lines = append(lines, fmt.Sprintf("%s\t%s", bestCdnAddr, twimgHostName))
	}
	
	fs.Seek(0, 0)
	fs.Truncate(0)
	writer := bufio.NewWriter(fs)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	writer.Flush()
	
	if patched {
		println("Twimg cdn 패치를 제거하였습니다.")
	} else {
		println("Twimg cdn 을 패치하였습니다.")
		println("최적의 cdn : " + bestCdnAddr)
	}
	
}

func readAllHosts(fs io.Reader) (lines []string, patched bool) {
	reader := bufio.NewReader(fs)

	befMatched := false

	for {
		lineBytes, _, err := reader.ReadLine()
		if err != nil && err != io.EOF {
			panic(err)
		}
		if lineBytes == nil {
			break
		}

		lineString := b2s(lineBytes)

		if befMatched {
			befMatched = false;
			if strings.TrimSpace(lineString) != "" {
				continue
			}
		}

		if regPattern.MatchString(lineString) {
			befMatched = true
			patched = true
		} else {
			lines = append(lines, lineString)
		}
	}

	return
}

