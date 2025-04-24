package models

import shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"

type Collection struct {
	DisplayName    string
	CollectionFile string
	Games          shared.Items
}

func (c Collection) Value() interface{} {
	return c
}
