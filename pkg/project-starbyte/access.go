package main

import (
    "fmt"
    "os"
)

type User struct{
    Name   string `json:"name"`
    accessLevel int `json:"accessLevel"`
    accessCode  int `json:"accessCode"`
}

func (a *User) accessControl(){
    jsonFile, err := os.Open("users.json")
        if err != nil {
            fmt.Println(err)
        }
    fmt.Println("Successfully Opened users.json")
    defer jsonFile.Close()
}


