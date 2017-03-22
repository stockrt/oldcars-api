package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

type CreateCarHandler struct {
	repo *CarRepository
}

func (h *CreateCarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	car := &Car{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(car)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.repo.Create(car)

	if err == ErrDuplicatedCar {
		fmt.Fprintln(w, "Carro já existe na base:", car)
	} else if err != nil {
		fmt.Fprintln(w, "Carro criado com sucesso:", car)
	}
}

type GetCarHandler struct {
	repo *CarRepository
}

func (h *GetCarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := httptreemux.ContextParams(r.Context())
	fmt.Fprintf(w, "Buscando carro com ID: %s", params["id"])

	car, err := h.repo.FindById(params["id"])

	if err == nil {
		fmt.Fprintln(w, "Carro:", car)
	} else {
		fmt.Fprintln(w, "Carro nao encontrado!")
	}
}

type ListAllCarsCarHandler struct {
	repo *CarRepository
}

func (h *ListAllCarsCarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cars, err := h.repo.ListAll()

	if err != nil {
		log.Println("Failed to fetch cars:", err)
	}

	fmt.Fprintln(w, "Lista de carros antigos:")

	for _, car := range cars {
		fmt.Fprintln(w, "- :", car)
	}

}

func main() {
	session, err := mgo.Dial("localhost:27017/oldcars")

	if err != nil {
		log.Fatal(err)
	}

	repository := NewCarRepository(session)

	addr := "127.0.0.1:8080"
	router := httptreemux.NewContextMux()
	router.Handler(http.MethodGet, "/cars", &ListAllCarsCarHandler{repository})
	router.Handler(http.MethodGet, "/cars/:id", &GetCarHandler{repository})
	router.Handler(http.MethodPut, "/cars", &CreateCarHandler{repository})

	log.Printf("Running web server on: http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

/*
	// updating a car
	car.Name = "Juliana updated"
	err = repository.Update(car)

	if err != nil {
		log.Println("Failed to update a car:", err)
	}

	repository.Create(&Car{Id: "124", Name: "Marcos"})
	repository.Create(&Car{Id: "125", Name: "Kaio", Inative: false})
	repository.Create(&Car{Id: "126", Name: "Gabriel"})
	repository.Create(&Car{Id: "127", Name: "Maisa"})

	// remove
	err = repository.Remove("126")

	if err != nil {
		log.Println("Failed to remove a car:", err)
	}

	// FindById
	car, err = repository.FindById("123")
	if err == nil {
		log.Printf("Result of findById: %v\n", car)
	} else {
		log.Println("Failed to findById", err)
	}
}
*/
