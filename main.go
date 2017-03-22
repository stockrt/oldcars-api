package main

import (
	"errors"
	"github.com/dimfeld/httptreemux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
)

const CarCollection = "car"

var ErrDuplicatedCar = errors.New("Duplicated car")

type Car struct {
	Id    string `bson:"_id"`
	Make  string `bson:"make"`
	Model string `bson:"model"`
	Year  int    `bson:"year"`
}

type CarRepository struct {
	session *mgo.Session
}

func (r *CarRepository) Create(p *Car) error {
	session := r.session.Clone()
	defer session.Close()

	collection := session.DB("").C(CarCollection)
	err := collection.Insert(p)
	if mongoErr, ok := err.(*mgo.LastError); ok {
		if mongoErr.Code == 11000 {
			return ErrDuplicatedCar
		}
	}
	return err
}

func (r *CarRepository) Update(p *Car) error {
	session := r.session.Clone()
	defer session.Close()

	collection := session.DB("").C(CarCollection)
	return collection.Update(bson.M{"_id": p.Id}, p)
}

func (r *CarRepository) Remove(id string) error {
	session := r.session.Clone()
	defer session.Close()

	collection := session.DB("").C(CarCollection)
	return collection.Remove(bson.M{"_id": id})
}

func (r *CarRepository) ListAll() ([]*Car, error) {
	session := r.session.Clone()
	defer session.Close()

	collection := session.DB("").C(CarCollection)
	query := bson.M{}

	documents := make([]*Car, 0)

	err := collection.Find(query).All(&documents)
	return documents, err
}

func (r *CarRepository) FindByYear() ([]*Car, error) {
	session := r.session.Clone()
	defer session.Close()

	collection := session.DB("").C(CarCollection)
	query := bson.M{"year": bson.M{"$gte": 1900}}

	documents := make([]*Car, 0)

	err := collection.Find(query).All(&documents)
	return documents, err
}

func (r *CarRepository) FindById(id string) (*Car, error) {
	session := r.session.Clone()
	defer session.Close()

	collection := session.DB("").C(CarCollection)
	query := bson.M{"_id": id}

	car := &Car{}

	err := collection.Find(query).One(car)
	return car, err
}

func NewCarRepository(session *mgo.Session) *CarRepository {
	return &CarRepository{session}
}

type GetCarHandler struct{}

func (h *GetCarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := httptreemux.ContextParams(r.Context())
	fmt.Fprintf(w, "Eu deveria busca um carro chamado: %s!", params["id"])
	fmt.Fprintln(w, "Não busco por que estou com preguiça!")
}

type ListAllCarsCarHandler struct{}

func (h *ListAllCarsCarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := httptreemux.ContextParams(r.Context())
	fmt.Fprintf(w, "Eu deveria busca um carro chamado: %s!", params["id"])
	fmt.Fprintln(w, "Não busco por que estou com preguiça!")
}

func main() {
	session, err := mgo.Dial("localhost:27017/oldcars")

	if err != nil {
		log.Fatal(err)
	}

	repository := NewCarRepository(session)

	addr := ":8080"
	router := httptreemux.NewContextMux()
	router.Handler(http.MethodGet, "/cars", &ListAllCarsCarHandler{})
	router.Handler(http.MethodGet, "/cars/:id", &GetCarHandler{})

	log.Printf("Running web server on: http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

/*
	// creating a car
	car := &Car{Id: "130", Name: "Juliana"}
	err = repository.Create(car)

	if err == ErrDuplicatedCar {
		log.Printf("%s is already created\n", car.Name)
	} else if err != nil {
		log.Println("Failed to create a car: ", err)
	}

	// updating a car
	car.Name = "Juliana updated"
	err = repository.Update(car)

	if err != nil {
		log.Println("Failed to update a car: ", err)
	}

	repository.Create(&Car{Id: "124", Name: "Marcos"})
	repository.Create(&Car{Id: "125", Name: "Kaio", Inative: false})
	repository.Create(&Car{Id: "126", Name: "Gabriel"})
	repository.Create(&Car{Id: "127", Name: "Maisa"})

	// remove
	err = repository.Remove("126")

	if err != nil {
		log.Println("Failed to remove a car: ", err)
	}

	// findAll
	people, err := repository.FindAllActive()
	if err != nil {
		log.Println("Failed to fetch people: ", err)
	}

	for _, car := range people {
		log.Printf("Have in database: %#v\n", car)
	}

	// FindById
	car, err = repository.FindById("123")
	if err == nil {
		log.Printf("Result of findById: %v\n", car)
	} else {
		log.Println("Failed to findById ", err)
	}
}
*/
