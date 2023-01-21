package tools

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/joho/godotenv"
)

func goDotEnvVariable(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	return os.Getenv(key)
}

func Contains(s []interface{}, value string) bool {
	for _, val := range s {
		if val == value {
			return true
		}
	}
	return false
}

func RemoveFromSlice(s []interface{}, value string) []interface{} {
	var index int
	for key, val := range s {
		if val == value {
			index = key
		}
	}
	sLastIndex := len(s) - 1
	if index != sLastIndex {
		s[index] = s[sLastIndex]
	}
	return s[:sLastIndex]
}

func ListDocumentSnapshots(s []*firestore.DocumentSnapshot, ch chan []interface{}) {
	var result []interface{}
	for _, val := range s {
		result = append(result, val.Data())
	}
	ch <- result
	return
}

func ListDocumentSnapshotsN(s []*firestore.DocumentSnapshot) []interface{} {
	var result []interface{}

	for _, val := range s {
		result = append(result, val.Data())
	}

	return result
}

func FindEntryInCollection(s *firestore.Client, collection string, entryId string, ch chan *firestore.DocumentSnapshot) {
	result, _ := s.Collection(collection).Doc(entryId).Get(context.TODO())

	ch <- result
	return
}

func FindAssociationsFromId(s *firestore.Client, collection string, field string, id string) []*firestore.DocumentSnapshot {
	result, _ := s.Collection(collection).Where(field, "==", id).Documents(context.TODO()).GetAll()
	return result
}

func AddDisputesToPost(s *firestore.Client, post *firestore.DocumentSnapshot, ch chan map[string]interface{}) {
	postDoc := post.Data()
	result, _ := s.Collection("disputes").Where("postId", "==", post.Ref.ID).Documents(context.TODO()).GetAll()
	if len(result) > 0 {
		postDoc["disputes"] = result[0].Data()
	}
	ch <- postDoc
}

func UserIdFromAuth(s *firestore.Client, authId string) []*firestore.DocumentSnapshot {
	result, _ := s.Collection("users").Where("authId", "==", authId).Documents(context.TODO()).GetAll()
	return result
}

func FindEntryInCollectionN(s *firestore.Client, collection string, entryId string) *firestore.DocumentSnapshot {
	result, _ := s.Collection(collection).Doc(entryId).Get(context.TODO())
	return result
}

func TestQuery1(s *firestore.Client, clubId string, authId string) ([]*firestore.DocumentSnapshot, *firestore.DocumentSnapshot) {
	creator, _ := s.Collection("users").Where("authId", "==", authId).Documents(context.TODO()).GetAll()
	club, _ := s.Collection("clubs").Doc(clubId).Get(context.TODO())

	return creator, club
}

func TestQuery3(s *firestore.Client, clubId string, authId string) ([]*firestore.DocumentSnapshot, *firestore.DocumentSnapshot) {

	creator := UserIdFromAuth(s, authId)
	club := FindEntryInCollectionN(s, "clubs", clubId)

	return creator, club
}
