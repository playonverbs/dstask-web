package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

	err = templs.ExecuteTemplate(rw, "index.html", ts)
	if err != nil {
		http.Error(rw, "failed to parse template", http.StatusInternalServerError)
		return
	}
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
			err := templs.ExecuteTemplate(rw, "task.html", t)
			if err != nil {
				http.Error(rw, "failed to parse template", http.StatusInternalServerError)
				return
			}
			return
		}
	}
}

func TaskTagHandler(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tags := strings.Split(vars["tag"], ",")

	ts, err := dstask.LoadTaskSet(conf.Repo, conf.IDsFile, false)
	if err != nil {
		http.Error(rw, "failed to load tasks", http.StatusInternalServerError)
		return
	}

	ts.Filter(dstask.Query{
		Tags: tags,
	})
	ts.SortByCreated(dstask.Ascending)
	ts.SortByPriority(dstask.Ascending)

	err = templs.ExecuteTemplate(rw, "tag.html", ts)
	if err != nil {
		http.Error(rw, "failed to parse template", http.StatusInternalServerError)
		return
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

func APIAddHandler(rw http.ResponseWriter, r *http.Request) {}

func APIProjectsHandler(rw http.ResponseWriter, r *http.Request) {
	ts, err := dstask.LoadTaskSet(conf.Repo, conf.IDsFile, false)
	if err != nil {
		http.Error(rw, "failed to load tasks", http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(ts.GetProjects())
	if err != nil {
		http.Error(rw, "failed to marshal to json", http.StatusInternalServerError)
		return
	}

	rw.Write(b)
	return
}
