package controllers

import (
	"fmt"
	//"log"
	"net/http"
	//"errors"
	"../models"
	"github.com/gorilla/mux"
	//"github.com/satori/go.uuid"
	"encoding/json"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type OrderController struct {
	session *mgo.Session
}

func NewOrderController(session *mgo.Session) *OrderController {
	return &OrderController{session}
}

func (oc OrderController) PlaceOrder(w http.ResponseWriter, req *http.Request) {
	var newOrder models.Order

	json.NewDecoder(req.Body).Decode(&newOrder)

	newOrder.ID = bson.NewObjectId()
	links := make(map[string]string)
	links["payment"] = "http://localhost:9090/starbucks/" + newOrder.ID.Hex() + "/pay"
	links["order"] = "http://localhost:9090/starbucks/order/" + newOrder.ID.Hex()
	newOrder.Links = links
	newOrder.Status = "payment expected"
	newOrder.Message = "Order has been placed."

	err := oc.session.DB("Restbucks").C("Order").Insert(&newOrder)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	fmt.Println("new order created: ", newOrder.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(newOrder)
}

func (oc OrderController) GetOrder(w http.ResponseWriter, req *http.Request) {
	//fmt.Println("here")
	params := mux.Vars(req)
	id := params["id"]
	//fmt.Println("id: ",id)

	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	oid := bson.ObjectIdHex(id)
	//fmt.Println("oid: ",oid)
	order := models.Order{}
	err := oc.session.DB("Restbucks").C("Order").FindId(oid).One(&order)

	if err != nil {
		//panic("abc")
		w.WriteHeader(404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(order)
	return
}

func (oc OrderController) CancelOrder(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	id := params["id"]
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	oid := bson.ObjectIdHex(id)
	order := models.Order{}
	err := oc.session.DB("Restbucks").C("Order").FindId(oid).One(&order)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	if order.Status != "payment expected" {
		w.WriteHeader(403)
		return
	}

	err = oc.session.DB("Restbucks").C("Order").RemoveId(oid)

	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(204)
}

func (oc OrderController) UpdateOrder(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	id := params["id"]
	//fmt.Println("id: ",id)

	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	oid := bson.ObjectIdHex(id)
	order := models.Order{}
	err := oc.session.DB("Restbucks").C("Order").FindId(oid).One(&order)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	if order.Status != "payment expected" {
		w.WriteHeader(403)
		return
	}

	json.NewDecoder(req.Body).Decode(&order)
	order.Message = "Order has been updated."

	err = oc.session.DB("Restbucks").C("Order").UpdateId(oid, order)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(order)
}

func (oc OrderController) PayOrder(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	id := params["id"]
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	oid := bson.ObjectIdHex(id)
	order := models.Order{}
	err := oc.session.DB("Restbucks").C("Order").FindId(oid).One(&order)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	if order.Status != "payment expected" {
		w.WriteHeader(403)
		return
	}

	order.Status = "PREPARING"
	order.Message = "Order has been paid."
	delete(order.Links, "payment")

	err = oc.session.DB("Restbucks").C("Order").UpdateId(oid, order)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(order)
}
