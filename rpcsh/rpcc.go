package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/rpc"
	"os"
	"path"
)

type Args struct {
	Id   int64
	Argv []string
}

type Result struct {
	Cmd    string
	Id     int64
	Stdout string
	Stderr string
	err    error
}

func readIps(fileName string) []string {
	dcFilter := os.ExpandEnv("${RPCSH_DC_FILTER}")
	regionFilter := os.ExpandEnv("${RPCSH_REGION_FILTER}")
	ips := make([]string, 0)
	if fh, err := os.Open(fileName); err == nil {
		defer fh.Close()
		r := csv.NewReader(fh)
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("error reading line in %s: %v", fileName, err)
			}

			if regionFilter != "" {
				if record[0] != regionFilter {
					continue
				}
			}

			if dcFilter != "" {
				if record[3] != dcFilter {
					continue
				}
			}

			ips = append(ips, record[2])
			fmt.Println(record[2])
		}
	}
	return ips
}

func main() {
	ipFile := os.ExpandEnv("${RPCSH_IP_FILE}")
	if ipFile == "" {
		ipFile = "/root/cloudian/survey.csv"
	}
	ips := readIps(ipFile)
	if len(ips) == 0 {
		fmt.Println("No ips found or no ips match your region or dc filter")
		return
	}

	prog := path.Base(os.Args[0])
	if prog == "all" {
		/*
		   args := Args{Message: *msg}
		   result := Result{}
		   serviceCall := client.Go("NetService.ToUpper", args, &result, nil)
		   select {
		   case <-time.After(time.Second * 1):
		       fmt.Println("Timeout")
		   case reply := <-serviceCall.Done:
		       if reply != nil {
		           if reply.Error == nil {
		               fmt.Println("Reply:", reply.Reply.(*Result).Result)
		           } else {
		               fmt.Fprintf(os.Stderr, "Err: %v\n", reply.Error)
		           }
		       } else {
		           fmt.Println("Error")
		       }
		   }
		*/
	}

	if prog == "all_wait" {
		for _, ip := range ips {
			client, err := rpc.Dial("tcp", fmt.Sprintf("%s:9999", ip))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}

			args := Args{}
			res := Result{}
			args.Id = rand.Int63()
			fmt.Fprintf(os.Stdout, "Starting command on %s with id %d\n", ip, args.Id)
			args.Argv = os.Args[1:]
			err = client.Call("CmdService.RunCommand", args, &res)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error running command on %s: %v", ip, err)
				return
			}
			fmt.Println(res.Id)
			fmt.Println(res.Stdout, "\n")
			fmt.Fprintf(os.Stderr, "%s\n\n", res.Stderr)
		}
	}
}