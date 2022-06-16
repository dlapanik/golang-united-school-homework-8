package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

type Arguments map[string]string

const (
	add      = "add"
	list     = "list"
	findById = "findById"
	remove   = "remove"
)

var (
	errNoOperationFlag = errors.New("-operation flag has to be specified")
	errNoFileNameFlag  = errors.New("-fileName flag has to be specified")
	errNoItemFlag      = errors.New("-item flag has to be specified")
	errNoIdFlag        = errors.New("-id flag has to be specified")

	operationArg string
	fileNameArg  string
	itemArg      string
	idArg        string

	allowedOperations = [...]string{add, list, findById, remove}
)

type User struct {
	ID    string `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
	Age   int    `json:"age,omitempty"`
}

func init() {
	flag.StringVar(&operationArg, "operation", "", "operation to be performed")
	flag.StringVar(&fileNameArg, "fileName", "", "file name with user data")
	flag.StringVar(&itemArg, "item", "", "item to be created")
	flag.StringVar(&idArg, "id", "", "person id")

	flag.Parse()
}

func Perform(args Arguments, writer io.Writer) error {
	err := validateArgs(args)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	switch args["operation"] {
	case list:
		fileData, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		writer.Write(fileData)

	case add:
		user, err := strToUser(args["item"])
		if err != nil {
			return err
		}

		users, err := readUsers(file)
		if err != nil {
			return err
		}

		users = append(users, user)

		err = saveUsers(file, users)
		if err != nil {
			return err
		}

	case findById:
		users, err := readUsers(file)
		if err != nil {
			return err
		}

		idToFind := args["id"]
		userFound := false
		for _, user := range users {
			if user.ID == idToFind {
				userRow, err := json.Marshal(user)
				if err != nil {
					return err
				}
				fmt.Fprintf(writer, string(userRow))
				break
			}
		}

		if !userFound {
			fmt.Fprintf(writer, "")
		}

	case remove:
		users, err := readUsers(file)
		if err != nil {
			return err
		}

		idToRemove := args["id"]
		userFound := false
		for i, user := range users {
			if user.ID == idToRemove {
				users = append(users[:i], users[i+1:]...)
				userFound = true
				saveUsers(file, users)
				break
			}
		}

		if !userFound {
			fmt.Fprintf(writer, "Item with id %s not found", idToRemove)
		}
	}
	return nil
}

func saveUsers(file *os.File, users []User) error {
	file.Truncate(0)
	file.Seek(0, 0)
	usersBytes, err := json.Marshal(users)
	if err != nil {
		return err
	}

	_, err = file.Write(usersBytes)

	return err
}

func readUsers(file *os.File) ([]User, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	items := []User{}

	if stat.Size() > 0 {
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&items)
		if err != nil {
			return nil, err
		}
	}

	return items, nil
}

func strToUser(itemAsStr string) (User, error) {
	item := User{}
	err := json.Unmarshal([]byte(itemAsStr), &item)
	return item, err
}

func validateArgs(args Arguments) error {
	// 1. validate file name
	if args["fileName"] == "" {
		return errNoFileNameFlag
	}

	// 2. validate operation flag and its values
	operation := args["operation"]
	if operation == "" {
		return errNoOperationFlag
	}

	correctOp := false
	for _, v := range allowedOperations {
		if operation == v {
			correctOp = true
			break
		}
	}

	if !correctOp {
		return fmt.Errorf("Operation %s not allowed!", args["operation"])
	}

	switch operation {
	case add:
		if args["item"] == "" {
			return errNoItemFlag
		}
	case remove, findById:
		if args["id"] == "" {
			return errNoIdFlag
		}
	}

	return nil
}

func parseArgs() Arguments {
	return Arguments{
		"id":        idArg,
		"operation": operationArg,
		"item":      itemArg,
		"fileName":  fileNameArg,
	}
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
