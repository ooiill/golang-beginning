package main

import (
    tool "beginning/pkg/acme"
    "os"
    "strings"
)

func main() {
    path := "/Users/echo/Downloads/poker"
    err := tool.DirectoryRecursion(path, func(file string, info os.FileInfo) error {
        saveFile := strings.ReplaceAll(file, "/poker", "/pk")
        _ = tool.MkDir(saveFile, true)

        err := tool.ResizeImage(file, saveFile, 200, 280)
        if err != nil {
            tool.PrintVar(err)
        }
        return nil
    }, nil)

    tool.PrintVar(err)
}
