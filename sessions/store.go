package sessions

import (
	"encoding/gob"
	"os"
)

type UserSession struct {
	Id string
}

type SessionId = string

type SessionStore interface {
	Save(UserSession) error
	Load(Id SessionId) (*UserSession, error)
}

type LocalSessionStore struct {
	Path string
}

func writeGob(filePath string, object interface{}) error {
	file, err := os.Create(filePath)
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(object)
	}
	file.Close()
	return err
}

func readGob(filePath string, object interface{}) error {
	file, err := os.Open(filePath)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	file.Close()
	return err
}

func (s LocalSessionStore) Save(us UserSession) error {
	return writeGob(s.Path+us.Id, us)
}

func (s LocalSessionStore) Load(id SessionId) (*UserSession, error) {
	userSession := UserSession{}
	err := readGob(s.Path+id, &userSession)
	if err != nil {
		return nil, err
	}
	return &userSession, nil
}
