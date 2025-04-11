package ui

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"strings"
)

func filterList(itemList []shared.Item, keywords ...string) []shared.Item {
	var filteredItemList []shared.Item

	for _, item := range itemList {
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(item.Filename), strings.ToLower(keyword)) {
				filteredItemList = append(filteredItemList, item)
				break
			}
		}
	}

	return filteredItemList
}
