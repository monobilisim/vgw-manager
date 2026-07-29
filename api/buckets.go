package api

import (
	"net/http"

	"github.com/omertahaoztop/vgw-manager/models"
	"github.com/omertahaoztop/vgw-manager/services"
)

// handleListBuckets returns the merged ZFS+API bucket list as JSON.
func handleListBuckets(w http.ResponseWriter, _ *http.Request) {
	buckets, err := services.ListMergedBuckets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, buckets)
}

// createBucketRequest is the JSON body for POST /v1/buckets.
type createBucketRequest struct {
	Name  string `json:"name"`
	Quota string `json:"quota"`
	Owner string `json:"owner"`
}

// handleCreateBucket creates a ZFS bucket and optionally sets its owner.
func handleCreateBucket(w http.ResponseWriter, r *http.Request) {
	var req createBucketRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Name == "" || req.Quota == "" {
		writeError(w, http.StatusBadRequest, errBucketNameQuotaRequired)
		return
	}

	bucketReq := models.BucketCreateRequest{
		Name:  req.Name,
		Quota: req.Quota,
		Owner: req.Owner,
	}
	bucketService := services.NewBucketService()
	if err := bucketService.CreateBucket(bucketReq); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if req.Owner != "" {
		vgwService := services.NewVersityGWService()
		if err := vgwService.ChangeBucketOwner(req.Name, req.Owner); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"name":  req.Name,
		"quota": req.Quota,
		"owner": req.Owner,
	})
}

// handleDeleteBucket deletes a bucket, trying ZFS first then API.
func handleDeleteBucket(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, errBucketNameRequired)
		return
	}

	via, err := services.DeleteBucketWithFallback(name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"bucket": name,
		"via":    via,
	})
}

// makePublicRequest is the optional JSON body for POST /v1/buckets/{name}/public.
type makePublicRequest struct {
	Owner string `json:"owner"`
}

// handleMakePublic makes a bucket public (read-only for everyone).
func handleMakePublic(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, errBucketNameRequired)
		return
	}

	var req makePublicRequest
	// Body is optional; ignore decode errors (empty body is fine).
	_ = decodeJSON(w, r, &req)

	if err := services.MakeBucketPublic(name, req.Owner); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"bucket": name,
		"status": "public",
	})
}

// handleMakePrivate removes the public policy from a bucket.
func handleMakePrivate(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, errBucketNameRequired)
		return
	}

	vgwService := services.NewVersityGWService()
	if err := vgwService.DeleteBucketPolicy(name); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"bucket": name,
		"status": "private",
	})
}
