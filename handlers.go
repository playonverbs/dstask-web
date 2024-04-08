package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/naggie/dstask"
)

func IndexHandler(rw http.ResponseWriter, r *http.Request) {
	ts, err := dstask.LoadTaskSet(conf.Repo, conf.IDsFile, false)
	if err != nil {
		http.Error(rw, "failed to load tasks", http.StatusInternalServerError)
		return
	}

	ts.SortByCreated(dstask.Ascending)
	ts.SortByPriority(dstask.Ascending)

	t, err := template.ParseFiles("template/index.html")
	if err != nil {
		http.Error(rw, "failed to parse template", http.StatusInternalServerError)
		return
	}

	t.Execute(rw, ts.Tasks())
}

func TaskIndexHandler(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	ts, err := dstask.LoadTaskSet(conf.Repo, conf.IDsFile, false)
	if err != nil {
		http.Error(rw, "failed to load tasks", http.StatusInternalServerError)
		return
	}

	for _, t := range ts.Tasks() {
		if t.UUID == uuid {
			templ, err := template.ParseFiles("template/task.html")
			if err != nil {
				http.Error(rw, "failed to parse template", http.StatusInternalServerError)
				return
			}

			templ.Execute(rw, t)
			return
		}
	}
}

func TaskAddHandler(rw http.ResponseWriter, r *http.Request) {
	ts, err := dstask.LoadTaskSet(conf.Repo, conf.IDsFile, false)
	if err != nil {
		http.Error(rw, "failed to load tasks", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(rw, "failed to parse form", http.StatusInternalServerError)
		return
	}

	desc := r.PostFormValue("desc")

	dstask.ParseQuery()
	t := dstask.Task{
		WritePending: true,
		Status:       dstask.STATUS_PENDING,
		Summary:      desc,
	}

	ch := make(chan error, 1)

	go func() {
		t, err = ts.LoadTask(t)
		if err != nil {
			ch <- err
			return
		}
		ts.SavePendingChanges()

		ch <- dstask.GitCommit(conf.Repo, "Added %s", t.Summary)
	}()

	if err = <-ch; err != nil {
		http.Error(rw, "failed to write task to disk", http.StatusInternalServerError)
		return
	}

	http.Redirect(rw, r, fmt.Sprintf("/task/%s", t.UUID), http.StatusFound)
	return
}

func APINextHandler(rw http.ResponseWriter, r *http.Request) {
	ts, err := dstask.LoadTaskSet(conf.Repo, conf.IDsFile, false)
	if err != nil {
		http.Error(rw, "failed to load tasks", http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(ts.Tasks())
	if err != nil {
		http.Error(rw, "failed to marshall to json", http.StatusInternalServerError)
		return
	}

	rw.Write(b)
}

func APITaskHandler(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(rw, "failed to load tasks", http.StatusInternalServerError)
		return
	}

	ts, err := dstask.LoadTaskSet(conf.Repo, conf.IDsFile, false)
	if err != nil {
		http.Error(rw, "failed to load tasks", http.StatusInternalServerError)
		return
	}

	for _, t := range ts.Tasks() {
		if t.ID == id {
			b, err := json.Marshal(t)
			if err != nil {
				http.Error(rw, "failed to marshall to json", http.StatusInternalServerError)
				return
			}

			rw.Write(b)
			return
		}
	}

	rw.WriteHeader(http.StatusNotFound)
}

func APIAddHandler(rw http.ResponseWriter, h *http.Request) {}
