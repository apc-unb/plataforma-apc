package student

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"strings"

	"github.com/apc-unb/apc-api/web/components/user"
	"github.com/apc-unb/apc-api/web/utils"

	"github.com/togatoga/goforces"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
)

// CreateStudents recieve a list of students
// Checks if that list is not null (can't insert null list)
// Insert each student individually in database
// @param	db				pointer to database
// @param	students 		list of students
// @param	databaseName	name of database
// @param	collectionName	name of collection
// @return 	error 			function error
// TODO : Insert all students at the same time (if possible)
func CreateStudents(db *mongo.Client, api *goforces.Client, students []StudentCreate, databaseName, collectionName string) ([]user.UserCredentials, error) {

	var studentsReturn []user.UserCredentials
	var singlestudent user.UserCredentials
	var pwd string
	var err error
	var mongoReturn *mongo.InsertOneResult

	if len(students) == 0 {
		return nil, nil
	}

	collection := db.Database(databaseName).Collection(collectionName)
	collectionLogin := db.Database(databaseName).Collection(collectionName + "_login")

	for _, student := range students {

		pwd = generateRandomPassword()

		if singlestudent.Password, err = utils.HashAndSalt([]byte(pwd)); err != nil {
			return nil, err
		}

		if mongoReturn, err = collection.InsertOne(context.TODO(), student); err != nil {
			return studentsReturn, err
		} else {
			singlestudent.ID = mongoReturn.InsertedID.(primitive.ObjectID)
		}

		singlestudent.Matricula = student.Matricula

		if _, err = collectionLogin.InsertOne(context.TODO(), singlestudent); err != nil {
			return nil, err
		}

		studentsReturn = append(studentsReturn, singlestudent)

	}

	return studentsReturn, nil
}

// CreateStudentsFile recieve csv file of students
// Call function that parse that file and return list o students
// Checks if that list is not null (can't insert null list)
// Insert each student individually in database
// @param	db				pointer to database
// @param	request 		byte array file
// @param	databaseName	name of database
// @param	collectionName	name of collection
// @return 	error 			function error
// TODO : Insert all students at the same time (if possible)
func CreateStudentsFile(db *mongo.Client, request string, databaseName, collectionName string) ([]user.UserCredentials, error) {

	var studentsReturn []user.UserCredentials
	var singlestudent user.UserCredentials
	var pwd string
	var err error
	var mongoReturn *mongo.InsertOneResult
	var students []StudentCreate

	if students, err = getStudentsFromFile(db, request); err != nil {
		return nil, err
	}

	if len(students) == 0 {
		return nil, nil
	}

	collection := db.Database(databaseName).Collection(collectionName)
	collectionLogin := db.Database(databaseName).Collection(collectionName + "_login")

	for _, student := range students {

		pwd = generateRandomPassword()

		if singlestudent.Password, err = utils.HashAndSalt([]byte(pwd)); err != nil {
			return nil, err
		}

		if mongoReturn, err = collection.InsertOne(context.TODO(), student); err != nil {
			return studentsReturn, err
		} else {
			singlestudent.ID = mongoReturn.InsertedID.(primitive.ObjectID)
		}

		singlestudent.Matricula = student.Matricula

		if _, err = collectionLogin.InsertOne(context.TODO(), singlestudent); err != nil {
			return nil, err
		}

		studentsReturn = append(studentsReturn, singlestudent)

	}

	return studentsReturn, nil
}

// GetStudents return list of all students from Database
// Get all students at the same time and store inside cursor
// Decode each student inside student class and append into students array
// @param	db				pointer to database
// @param	databaseName	name of database
// @param	collectionName	name of collection
// @return 	[]Student		list of all students
// @return 	error 			function error
func GetStudents(db *mongo.Client, databaseName, collectionName string) ([]StudentInfo, error) {

	collection := db.Database(databaseName).Collection(collectionName)

	// Here's an array in which you can store the decoded documents
	students := []StudentInfo{}

	// Passing bson.D{{}} as the filter matches all documents in the collection
	cursor, err := collection.Find(
		context.TODO(),
		bson.D{{}},
		options.Find(),
	)

	if err != nil {
		return nil, err
	}

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cursor.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var elem StudentInfo

		// Checks if decoding method didn't return any errors
		if err := cursor.Decode(&elem); err != nil {
			return nil, err
		}

		// Push student inside student array
		students = append(students, elem)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	// Close the cursor once finished
	cursor.Close(context.TODO())

	return students, nil
}

// GetStudentsClass return list of all students from Database thata matchs with a certain class
// Get all students at the same time and store inside cursor
// Decode each student inside student class and append into students array
// @param	db				pointer to database
// @param   classID         ID of the current class
// @param	databaseName	name of database
// @param	collectionName	name of collection
// @return 	[]Student		list of all students
// @return 	error 			function error
func GetStudentsClass(db *mongo.Client, classID primitive.ObjectID, databaseName, collectionName string) ([]Student, error) {

	collection := db.Database(databaseName).Collection(collectionName)

	// Here's an array in which you can store the decoded documents
	students := []Student{}

	var options options.FindOptions

	options.SetSort(bson.D{{"firstname", 1}, {"lastname", 1}})

	// Passing bson.D{{}} as the filter matches all documents in the collection
	cursor, err := collection.Find(
		context.TODO(),
		bson.M{
			"classid": classID,
		},
		&options,
	)

	if err != nil {
		return nil, err
	}

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cursor.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var elem Student

		// Checks if decoding method didn't return any errors
		if err := cursor.Decode(&elem); err != nil {
			return nil, err
		}

		// Push school class inside student array
		students = append(students, elem)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	// Close the cursor once finished
	cursor.Close(context.TODO())

	return students, nil

}

func GetStudent(db *mongo.Client, studentID primitive.ObjectID, databaseName, collectionName string) (Student, error) {

	collection := db.Database(databaseName).Collection(collectionName)

	studentDAO := Student{}

	filter := bson.D{
		{"_id", studentID},
	}

	if err := collection.FindOne(
		context.TODO(),
		filter,
	).Decode(&studentDAO); err != nil {
		return studentDAO, err
	}

	return studentDAO, nil
}

// UpdateStudents recieve student (updated)
// Checks if student old password matches with db to update that student password or email
// @param	db				pointer to database (updated)
// @param	students 		list of students
// @param	databaseName	name of database
// @param	collectionName	name of collection
// @return 	StudentUpdate	student new data
// @return 	error 			function error
// TODO : Update all students at the same time (if possible)
func UpdateStudents(db *mongo.Client, api *goforces.Client, student StudentUpdate, databaseName, collectionName string) error {

	var err error
	collection := db.Database(databaseName).Collection(collectionName)
	collectionLogin := db.Database(databaseName).Collection(collectionName + "_login")

	studentData := user.UserCredentials{}
	currentStudent := Student{}
	update := bson.M{}

	filter := bson.M{
		"_id": student.ID,
	}

	if err := collectionLogin.FindOne(
		context.TODO(),
		filter,
		options.FindOne(),
	).Decode(&studentData); err != nil {
		return err
	}

	if err = utils.ComparePasswords(studentData.Password, student.Password); err != nil {
		return errors.New("mongo: no documents in result")
	}

	if student.NewPassword != "" {
		if student.NewPassword, err = utils.HashAndSalt([]byte(student.NewPassword)); err != nil {
			return err
		}

		updateSet := bson.M{"$set": bson.M{"password": student.NewPassword}}

		if _, err := collectionLogin.UpdateOne(context.TODO(), filter, updateSet, nil); err != nil {
			return err
		}

		return nil
	}

	projection := bson.M{
		"_id":     1,
		"handles": 1,
	}

	if err := collection.FindOne(
		context.TODO(),
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&currentStudent); err != nil {
		return err
	}

	if student.Email != "" {
		update["email"] = student.Email
	}

	if student.Handles.Codeforces != "" {
		if currentStudent.Handles.Codeforces != "" {
			return errors.New("Trying to update handle that already exist")
		} else {
			update["photourl"] = getCodeforcesAvatarURL(student.Handles.Codeforces, api)
			update["handles.codeforces"] = student.Handles.Codeforces
		}
	}

	if student.Handles.Uri != "" {
		if currentStudent.Handles.Uri != "" {
			return errors.New("Trying to update handle that already exist")
		} else {
			update["handles.uri"] = student.Handles.Uri
		}
	}

	updateSet := bson.M{"$set": update}

	if _, err := collection.UpdateOne(context.TODO(), filter, updateSet, nil); err != nil {
		return err
	}

	return nil
}

// DeleteStudents recieve a list of students (to be deleted)
// Checks if that list is not null (can't delete null list)
// Delete each student individually
// @param	db				pointer to database (to be deleted)
// @param	students 		list of students
// @param	databaseName	name of database
// @param	collectionName	name of collection
// @return 	[]Student		list of all students
// @return 	error 			function error
// TODO : Delete all students at the same time (if possible)
func DeleteStudents(db *mongo.Client, students []Student, databaseName, collectionName string) error {

	if len(students) == 0 {
		return nil
	}

	collection := db.Database(databaseName).Collection(collectionName)

	for _, student := range students {
		filter := bson.M{"_id": student.ID}
		if _, err := collection.DeleteOne(context.TODO(), filter); err != nil {
			return err
		}
	}
	return nil

}

// AuthStudent recieve a student (to be authenticated)
// Checks if that date exist in databse
// Return true if exist
// @param	db				pointer to database (to be deleted)
// @param	student			student matricula and password
// @param	databaseName	name of database
// @param	collectionName	name of collection
// @return 	[]bool			user exist veredict
// @return 	error 			function error
func AuthStudent(db *mongo.Client, student user.UserCredentials, databaseName, collectionName string) (StudentInfo, error) {

	var err error

	collection := db.Database(databaseName).Collection(collectionName)
	collectionLogin := db.Database(databaseName).Collection(collectionName + "_login")

	findStudent := StudentInfo{}
	studentData := user.UserCredentials{}

	if err := collectionLogin.FindOne(
		context.TODO(),
		bson.D{{"matricula", student.Matricula}},
		options.FindOne(),
	).Decode(&studentData); err != nil {
		return findStudent, err
	}

	if err = utils.ComparePasswords(studentData.Password, student.Password); err != nil {
		return findStudent, errors.New("mongo: no documents in result")
	}

	if err := collection.FindOne(
		context.TODO(),
		bson.D{{"_id", studentData.ID}},
		options.FindOne(),
	).Decode(&findStudent); err != nil {
		return findStudent, err
	}

	return findStudent, nil
}

// getCodeforcesAvatarURL recieve handle string
// Return handle avatar url if exist
// @param	handle			student codeforces handle
// @param	api				pointer to goforces client
// @return 	string 			avatar url
func getCodeforcesAvatarURL(handle string, api *goforces.Client) string {

	ctx := context.Background()

	var userAvatarURL string

	if handlesArray, err := api.GetUserInfo(ctx, []string{handle}); err == nil {
		userAvatarURL = "https:" + handlesArray[0].Avatar
	}

	return userAvatarURL
}

// getClassID recieve class year, season and class name
// Return current class id
// @param	db				pointer to database (to be deleted)
// @param	classData		year, season and class name
// @param	databaseName	name of database
// @param	collectionName	name of collection
// @return 	[]ObjectID		class id
// @return 	error 			function error
func getClassID(db *mongo.Client, classData []string, databaseName, collectionName string) (primitive.ObjectID, error) {

	type teste struct {
		ID primitive.ObjectID `bson:"_id,omitempty"`
	}

	var classID teste

	if len(classData) != 3 {
		return classID.ID, errors.New("YEAR/SEASON/CLASSNAME header error")
	}

	year, _ := strconv.Atoi(strings.Trim(classData[0], "\""))
	season, _ := strconv.Atoi(classData[1])
	classname := strings.Trim(classData[2], "\"")

	collection := db.Database(databaseName).Collection(collectionName)

	filter := bson.D{
		{"year", year},
		{"season", season},
		{"classname", classname},
	}

	projection := bson.D{
		{"_id", 1},
	}

	if err := collection.FindOne(
		context.TODO(),
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&classID); err != nil {
		return classID.ID, err
	}

	return classID.ID, nil
}

// generateRandomPassword return
// Return current class id
// @param	db				pointer to database (to be deleted)
// @param	classData		year, season and class name
// @param	databaseName	name of database
// @param	collectionName	name of collection
// @return 	[]ObjectID		class id
// @return 	error 			function error
func generateRandomPassword() string {

	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"
	var password string

	for i := 0; i < 3; i++ {
		password += string(letters[rand.Intn(25)])
	}
	for i := 0; i < 3; i++ {
		password += string(numbers[rand.Intn(9)])
	}

	return password

}

func getStudentsFromFile(db *mongo.Client, request string) ([]StudentCreate, error) {

	var total []string
	var students []StudentCreate
	var classID primitive.ObjectID
	var err error

	partial := strings.Split(request, ",")
	total = append(total, strings.Split(partial[1], "\n")[1])
	classData := strings.Split(strings.Split(partial[1], "\n")[0], "/")

	if classID, err = getClassID(db, classData, "apc_database", "schoolClass"); err != nil {
		return students, err
	}

	for i := 2; i < len(partial); i++ {
		aux := strings.Split(partial[i], "\n")
		total = append(total, aux[0])
		total = append(total, aux[1])
	}
	total = total[:len(total)-1]

	for i := 0; i < len(total); i += 2 {

		names := strings.SplitAfterN(total[i+1], " ", 2)

		elem := StudentCreate{

			FirstName: strings.Trim(names[0], "\""),
			LastName:  strings.Trim(names[1], "\""),
			Matricula: strings.Trim(total[i], "\""),
			ClassID:   classID,
			//Password:  generateRandomPassword(),
		}

		students = append(students, elem)
	}

	return students, nil
}

func GetUserProgress (contestsIds []int, handle string, api *goforces.Client) (interface{}, error){

	ctx := context.Background()
	done := 0
	total := 0
	opt := goforces.ContestStatndingsOptions{
		Handles:        []string{handle},
		ShowUnofficial: true,
	}

	for i := 0; i < len(contestsIds); i++ {
		standings, err := api.GetContestStandings(ctx, contestsIds[i], &opt)

		if err != nil {
			return nil, err
		}

		totalProblems := len(standings.Problems)
		total += totalProblems
		tasks := make([]int, totalProblems)

		for j := 0; j < len(standings.Rows); j++ {
			for k := 0; k < totalProblems; k++ {
				score := int(standings.Rows[j].ProblemResults[k].Points)
				done += (score ^ tasks[k]) & (tasks[k] ^ 1)
				tasks[k] |= score
			}
		}

	}

	userProgress := map[string]interface{}{
		"done": strconv.Itoa(done),
		"total": strconv.Itoa(total),
	}

	return userProgress, nil

}

func GetIndividualUserProgress (contestsIds []int, handle, groupID string, api *goforces.Client) (interface{}, error){
	ctx := context.Background()
	opt := goforces.ContestStatndingsOptions{
		Handles:        []string{handle},
		ShowUnofficial: true,
	}

	var contests []interface{}

	for i := 0; i < len(contestsIds); i++ {

		standings, err := api.GetContestStandings(ctx, contestsIds[i], &opt)

		if err != nil {
			return nil, err
		}

		total := len(standings.Problems)
		done := 0
		tasks := make([]int, total)

		for j := 0; j < len(standings.Rows); j++ {
			for k := 0; k < total; k++ {
				score := int(standings.Rows[j].ProblemResults[k].Points)
				done += (score ^ tasks[k]) & (tasks[k] ^ 1)
				tasks[k] |= score
			}
		}

		userProgress := map[string]interface{}{
			"name": standings.Contest.Name,
			"url": "https://codeforces.com/group/" + groupID + "/contest/" + strconv.Itoa(int(standings.Contest.ID)),
			"done" : done,
			"total" : total,
		}

		contests = append(contests, userProgress)
	}

	return contests, nil
}