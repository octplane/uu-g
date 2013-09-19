package uu

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"os"
)

func savePost(params map[string]string) string {
	fname, mnem := res.pasteResolver.GetNextIdentifier()
	file, err := os.OpenFile(fname, os.O_EXCL|os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var count int
	var data []byte

	var paste = make(map[string]interface{})

	paste["content"] = params["content"]
	paste["attachments"] = params["attachments"]
	paste["expire"] = makeExpiryFromPost(params["expiry_delay"], params["never_expire"] == "true")

	data, err = json.Marshal(paste)
	if err != nil {
		panic(err)
	}

	count, err = file.Write(data)
	if err != nil {
		panic(err)
	}
	if count != len(data) {
		panic(fmt.Sprintf("Wrote only %d/%d in %s", count, len(data), fname))
	}

	return mnem
}

func saveAttachment(attn multipart.File, prefix string) string {

	content, err := ioutil.ReadAll(attn)

	if err != nil {
		panic(err)
	}

	fname, mnem := res.attnResolver.GetNextIdentifierWithPrefix(prefix)
	file, err := os.OpenFile(fname, os.O_EXCL|os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var count int
	count, err = file.Write(content)
	if err != nil {
		panic(err)
	}
	if count != len(content) {
		panic(fmt.Sprintf("Wrote only %d/%d in %s", count, len(content), fname))
	}
	return mnem
}
