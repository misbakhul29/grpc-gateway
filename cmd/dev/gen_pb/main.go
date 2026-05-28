package main

import (
	"grpc/gateway/internal/logger"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	logger.Log.Info("gen_pb", "Generating protobuf files")
	files, err := filepath.Glob("proto/*.proto")
	if err != nil {
		logger.Log.Fatal("gen_pb", err.Error())
	}
	if len(files) == 0 {
		logger.Log.Warn("gen_pb", "No protobuf files found")
		return
	}
	cmdList := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/grpc-ecosystem/grpc-gateway/v2")
	gatewayDirBytes, err := cmdList.Output()
	if err != nil {
		logger.Log.Fatal("gen_pb", err.Error())
	}
	gatewayDir := strings.TrimSpace(string(gatewayDirBytes))

	args := append([]string{
		"-I.",
		"-Ithird_party",
		"-I" + gatewayDir,
		"--go_out=.",
		"--go-grpc_out=.",
		"--grpc-gateway_out=.",
		"--openapiv2_out=allow_merge=true,merge_file_name=docs:proto",
	}, files...)

	cmd := exec.Command("protoc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logger.Log.Fatal("gen_pb", err.Error())
	}
	oldPath := filepath.Join("proto", "docs.swagger.json")
	newPath := filepath.Join("proto", "docs.json")
	if _, err := os.Stat(oldPath); err == nil {
		_ = os.Remove(newPath)
		if err := os.Rename(oldPath, newPath); err != nil {
			logger.Log.Fatal("gen_pb", err.Error())
		}
	}
	logger.Log.Info("gen_pb", "Protobuf files generated successfully in ./pb/")
}
