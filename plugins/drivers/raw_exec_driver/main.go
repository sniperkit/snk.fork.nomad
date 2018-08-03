package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-plugin"

	"strings"

	"bufio"

	"github.com/golang/protobuf/jsonpb"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/hashicorp/nomad/plugins/drivers/raw_exec_driver/proto"
	"github.com/hashicorp/nomad/plugins/drivers/raw_exec_driver/shared"
)

func main() {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         shared.PluginMap,
		Cmd:             exec.Command("sh", "-c", "./raw-exec-go-grpc"),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolGRPC},
	})
	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error when trying to start rpc client:", err.Error())
		os.Exit(1)
	}

	raw, err := rpcClient.Dispense("raw_exec")
	if err != nil {
		fmt.Println("Error when dispensing raw_exec:", err.Error())
		os.Exit(1)
	} else if raw == nil {
		fmt.Println("Error when dispensing raw_exec: got null instead of interface")
		os.Exit(1)
	}

	rawExec := raw.(shared.RawExec)

	args := os.Args[1:]
	fmt.Println(args)
	if args[0] == "start" {
		doStart(rawExec)
	} else if args[0] == "restore" {
		doRestore(rawExec)
	}

}
func doStart(rawExec shared.RawExec) {
	fmt.Println("Starting...")
	currentDir, err := os.Getwd()
	// TODO
	if err != nil {
		panic(fmt.Sprintf("encountered error when getting current dir: %s", err.Error()))
	}
	execCtx := &proto.ExecContext{
		TaskDir: &proto.TaskDir{
			Directory: currentDir,
			LogDir:    currentDir,
			LogLevel:  "DEBUG",
		},
		TaskEnv: &proto.TaskEnv{},
	}
	jsonConfig := `{
                     "Command":"bash",
                     "Args":["-c", "sleep 10000"]
                   }`
	unMarshaller := jsonpb.Unmarshaler{AllowUnknownFields: false}
	reader := strings.NewReader(jsonConfig)
	structConfig := &_struct.Struct{}
	err = unMarshaller.Unmarshal(reader, structConfig)
	if err != nil {
		fmt.Println("Error unmarshalling json into protobuf Struct:%v", err)
		os.Exit(-1)
	}
	taskInfo := &proto.TaskInfo{
		Resources: &proto.Resources{
			Cpu:      250,
			MemoryMb: 256,
			DiskMb:   20,
		},
		LogConfig: &proto.LogConfig{
			MaxFiles:      10,
			MaxFileSizeMb: 10,
		},
		Config: structConfig,
	}

	fmt.Println("Calling new start **")
	result, err := rawExec.NewStart(execCtx, taskInfo)
	if err != nil {
		fmt.Printf("Encountered errors: %s \n", err.Error())
	}
	fmt.Printf("After start: %+v \n", result)
	fmt.Printf("Serializing task state..")
	ts := result.TaskState
	marshaller := jsonpb.Marshaler{OrigName: true}
	f, err := os.Create(fmt.Sprintf("%v/%v.json", currentDir, ts.TaskId))
	defer f.Close()
	w := bufio.NewWriter(f)
	marshaller.Marshal(w, ts)
	w.Flush()
}

func doRestore(rawExec shared.RawExec) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Unexpected error getting current working directory:%v", err)
		return
	}
	taskStateFile := fmt.Sprintf("%v/%v.json", wd, "bc2bfd9f-ca91-3e91-343b-108c168862d3")
	f, err := os.Open(taskStateFile)
	defer f.Close()
	r := bufio.NewReader(f)
	unMarshaller := jsonpb.Unmarshaler{AllowUnknownFields: false}
	taskState := &proto.TaskState{}
	unMarshaller.Unmarshal(r, taskState)

	fmt.Printf("After unmarshalling json %+v\n", taskState)

	restoreResp, err := rawExec.Restore([]*proto.TaskState{taskState})
	if err != nil {
		fmt.Printf("Error in restoring state:%v\n", restoreResp)
		return
	}
	for _, resp := range restoreResp.RestoreResults {
		if resp.ErrorMessage != "" {
			fmt.Println("Error restoring task ", resp.TaskId)
		} else {
			fmt.Println("Successfully reattached to task ", resp.TaskId)
		}
	}
}
