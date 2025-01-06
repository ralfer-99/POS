package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection

func InitUserCollection(client *mongo.Client, dbName, collectionName string) {
	userCollection = client.Database(dbName).Collection(collectionName)
}

func AddToCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var reqBody struct {
		UserID string `json:"userId"`
		ItemID string `json:"itemId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := primitive.ObjectIDFromHex(reqBody.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var userData struct {
		CartData map[string]int `bson:"cartData"`
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&userData)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if userData.CartData == nil {
		userData.CartData = make(map[string]int)
	}
	userData.CartData[reqBody.ItemID]++

	_, err = userCollection.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{"cartData": userData.CartData}})
	if err != nil {
		http.Error(w, "Failed to update cart", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Added To Cart",
	})
}

func RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var reqBody struct {
		UserID string `json:"userId"`
		ItemID string `json:"itemId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := primitive.ObjectIDFromHex(reqBody.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var userData struct {
		CartData map[string]int `bson:"cartData"`
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&userData)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if userData.CartData != nil && userData.CartData[reqBody.ItemID] > 0 {
		userData.CartData[reqBody.ItemID]--

		if userData.CartData[reqBody.ItemID] == 0 {
			delete(userData.CartData, reqBody.ItemID)
		}

		_, err = userCollection.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{"cartData": userData.CartData}})
		if err != nil {
			http.Error(w, "Failed to update cart", http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Removed From Cart",
	})
}

func GetCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var reqBody struct {
		UserID string `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := primitive.ObjectIDFromHex(reqBody.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var userData struct {
		CartData map[string]int `bson:"cartData"`
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&userData)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"cartData": userData.CartData,
	})
}
