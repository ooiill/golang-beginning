package main

import (
    tool "beginning/pkg/acme"
    "io/ioutil"
    "os"
    "strings"
)

func main() {
    path := "/Users/echo/Downloads/mobile_svg_b"
    err := tool.DirectoryRecursion(path, func(file string, info os.FileInfo) error {
        xml, _ := ioutil.ReadFile(file)
        xmlString := string(xml)
        xmlString = strings.ReplaceAll(xmlString, "<rect x=\"-240\" y=\"-336\" width=\"480\" height=\"672\" fill=\"#fcfcfc\"></rect>", "")
        xmlString = strings.ReplaceAll(xmlString, "fill=\"#fcfcfc\" stroke=\"black\"></rect>", "fill=\"#fcfcfc\" stroke=\"white\"></rect>")
        err := ioutil.WriteFile(file, []byte(xmlString), 0777)
        if err != nil {
            tool.PrintVar(err)
        }
        return nil
    }, nil)
    tool.PrintVar(err)
}
