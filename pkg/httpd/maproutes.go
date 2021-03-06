// Copyright (c) 2017, CodeBoy. All rights reserved.
//
// This Source Code Form is subject to the terms of the
// license that can be found in the LICENSE file.

package httpd

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
	"time"

	"github.com/filemaps/filemaps/pkg/model"
)

func routeMaps(r *httprouter.Router) {
	mapsURL := APIURL + "/maps"
	r.GET(mapsURL, ReadMaps)
	r.POST(mapsURL, CreateMap)

	mapURL := mapsURL + "/:mapid"
	r.GET(mapURL, ReadMap)
	r.PUT(mapURL, UpdateMap)
	r.DELETE(mapURL, DeleteMap)
	r.POST(mapURL, ImportMap)

	routeResources(r, mapURL)
}

// ReadMaps is controller for getting maps.
// Returns array of all maps, sorted by Opened field.
func ReadMaps(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	mm := model.GetMapManager()
	resp := make(map[string]interface{})
	resp["maps"] = mm.GetMaps()
	WriteJSON(w, resp)
}

// CreateMap creates new Map.
func CreateMap(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	type JSONRequest struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Base        string `json:"base"`
		File        string `json:"file"`
		Template    string `json:"template"`
	}
	var jr JSONRequest
	d := json.NewDecoder(r.Body)
	err := d.Decode(&jr)
	r.Body.Close()
	if err != nil {
		WriteJSONError(w, 400, "bad request")
		return
	}

	log.WithFields(log.Fields{
		"title":       jr.Title,
		"description": jr.Description,
		"base":        jr.Base,
		"file":        jr.File,
		"template":    jr.Template,
	}).Info("Create Map")

	info := model.MapInfo{
		Title:  jr.Title,
		Base:   jr.Base,
		File:   jr.File,
		Opened: time.Now(),
	}
	mm := model.GetMapManager()
	pm, err := mm.AddMap(info)
	if err != nil {
		WriteJSONError(w, 500, "could not add map")
		return
	}
	pm.Description = jr.Description
	if err = pm.Write(); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Could not write map")
		WriteJSONError(w, 500, "could not add map")
		return
	}
	mm.Write()
	writeMap(w, pm)
}

// ImportMap imports existing Map.
func ImportMap(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if ps.ByName("mapid") != "import" {
		WriteJSONError(w, 400, "bad request")
		return
	}

	type JSONRequest struct {
		Path string `json:"path"`
	}
	var jr JSONRequest
	d := json.NewDecoder(r.Body)
	err := d.Decode(&jr)
	r.Body.Close()
	if err != nil {
		WriteJSONError(w, 400, "bad request")
		return
	}

	log.WithFields(log.Fields{
		"path": jr.Path,
	}).Info("Import Map")

	mm := model.GetMapManager()
	pm, err := mm.ImportMap(jr.Path)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Could not import map")
		WriteJSONError(w, 500, "could not import map")
		return
	}
	mm.Write()
	writeMap(w, pm)
}

// ReadMap is controller for getting a map.
func ReadMap(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("mapid"))
	if err != nil {
		WriteJSONError(w, 400, "map id must be integer")
		return
	}
	mm := model.GetMapManager()
	pm := mm.GetProxyMap(id)
	writeMap(w, pm)
}

// UpdateMap updates existing Map.
func UpdateMap(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type JSONRequest struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Base        string   `json:"base"`
		File        string   `json:"file"`
		Exclude     []string `json:"exclude"`
	}
	var jr JSONRequest
	d := json.NewDecoder(r.Body)
	err := d.Decode(&jr)
	r.Body.Close()
	if err != nil {
		WriteJSONError(w, 400, "bad request")
		return
	}

	log.WithFields(log.Fields{
		"title":       jr.Title,
		"description": jr.Description,
		"base":        jr.Base,
		"file":        jr.File,
		"exclude":     jr.Exclude,
	}).Info("Update Map")

	pm := findProxyMap(ps.ByName("mapid"))
	if pm == nil {
		WriteJSONError(w, 404, "map not found")
		return
	}

	mm := model.GetMapManager()
	mm.UpdateMap(pm.ID, jr.Title, jr.Description, jr.Base, jr.File, jr.Exclude)
	pm.Write()
	mm.Write()

	writeMap(w, pm)
}

// DeleteMap is controller for deleting a map.
func DeleteMap(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("mapid"))
	if err != nil {
		WriteJSONError(w, 404, "map not found")
		return
	}

	mm := model.GetMapManager()
	mm.DeleteMap(id)
	mm.Write()
	fmt.Fprint(w, "{}")
}

// findMap returns ProxyMap by ID or nil if not found.
func findProxyMap(param string) *model.ProxyMap {
	mapID, err := strconv.Atoi(param)
	if err != nil {
		return nil
	}

	mm := model.GetMapManager()
	pm := mm.GetProxyMap(mapID)
	if pm == nil {
		return nil
	}

	return pm
}

// writeMap writes ProxyMap to JSON response.
func writeMap(w http.ResponseWriter, pm *model.ProxyMap) {
	if pm != nil {
		pm.Read()
		m := pm.Map

		resp := make(map[string]interface{})
		resp["fileMap"] = m
		resp["defaultStyles"] = model.NewDefaultStyles()
		WriteJSON(w, resp)
	} else {
		WriteJSONError(w, 404, "map not found")
	}
}
