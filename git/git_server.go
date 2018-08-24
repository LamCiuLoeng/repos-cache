package git

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// GitServer 是 git 的 HTTP 服务器类
type GitServer struct {
	listenAddr string
	gitUtil    GitUtil
}

func (g GitServer) getRemoteRepoURL(r *http.Request) string {
	path := r.URL.Path[1:]
	suffixs := []string{"/info/refs", "/HEAD", "/git-upload-pack", "/git-receive-pack"}
	for _, suffix := range suffixs {
		if strings.HasSuffix(path, suffix) {
			path = path[:len(path)-len(suffix)]
		}
	}
	if !(strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")) {
		path = "https://" + path
	}
	return path
}

func (g GitServer) handleInfoRefs(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Query().Get("service")
	remoteRepoURL := g.getRemoteRepoURL(r)
	localRepoPath := g.gitUtil.FindLocalRepoPath(remoteRepoURL)

	g.gitUtil.CloneIfNotExist(remoteRepoURL)

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", serviceName))
	cmd := GitCommand{Args: []string{serviceName, "--stateless-rpc", "--advertise-refs", localRepoPath}}
	stdout, err := cmd.Run(false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	} else {
		str := "# service=" + serviceName
		fmt.Fprintf(w, "%.4x%s\n", len(str)+5, str)
		fmt.Fprintf(w, "0000")
		nbytes, err := io.Copy(w, stdout)
		if err != nil {
			log.Printf("执行命令失败 %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			log.Printf("Bytes written: %d", nbytes)
		}
	}
}

func (g GitServer) handleGitService(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Query().Get("service")
	if serviceName == "" {
		if strings.HasSuffix(r.URL.Path, "/git-upload-pack") {
			serviceName = "git-upload-pack"
		} else if strings.HasSuffix(r.URL.Path, "/git-receive-pack") {
			serviceName = "git-receive-pack"
		}
	}

	if serviceName == "" {
		w.WriteHeader(http.StatusForbidden)
		log.Println("收到异常请求，serverName 是空")
		return
	}
	remoteRepoURL := g.getRemoteRepoURL(r)
	localRepoPath := g.gitUtil.FindLocalRepoPath(remoteRepoURL)

	g.gitUtil.CloneIfNotExist(remoteRepoURL)
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-result", serviceName))
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Fatal("Error:", err)
		return
	}
	cmd := GitCommand{
		ProcInput: bytes.NewReader(requestBody),
		Args:      []string{serviceName, "--stateless-rpc", localRepoPath}}

	stdout, err := cmd.Run(false)
	if err != nil {
		log.Printf("执行命令失败 %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		gzw := gzip.NewWriter(w)
		defer gzw.Close()
		wrapper := gzipResponseWriter{Writer: gzw, ResponseWriter: w}
		nbytes, err := io.Copy(wrapper, stdout)
		if err != nil {
			log.Printf("gzip 写入失败 %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			log.Printf("gzip 写入字节 %d", nbytes)
			w.Header().Set("Content-Encoding", "gzip")
		}
	} else {
		nbytes, err := io.Copy(w, stdout)
		if err != nil {
			log.Printf("写入失败 %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			log.Printf("写入字节 %d", nbytes)
		}
	}

}

// Dispatcher 是 URL 分发器
func (g GitServer) Dispatcher(w http.ResponseWriter, r *http.Request) {
	log.Printf("=> %s %s\n", r.Method, r.URL)
	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == http.MethodGet {
		serviceName := r.URL.Query().Get("service")
		if strings.HasSuffix(r.URL.Path, "/info/refs") {
			g.handleInfoRefs(w, r)
		} else if serviceName == "git-receive-pack" || serviceName == "git-upload-pack" {
			g.handleGitService(w, r)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	} else if r.Method == http.MethodPost {
		if strings.HasSuffix(r.URL.Path, "/git-receive-pack") || strings.HasSuffix(r.URL.Path, "/git-upload-pack") {
			g.handleGitService(w, r)
		} else {
			w.WriteHeader(http.StatusNotImplemented)
		}
	}
}

// Run 运行服务器
func (g GitServer) Run() {
	http.HandleFunc("/", g.Dispatcher)
	log.Println("服务器运行在 " + g.listenAddr)
	log.Fatal(http.ListenAndServe(g.listenAddr, nil))
}

// NewGitServer 创建 git 的 HTTP 服务器
func NewGitServer(c *viper.Viper) GitServer {
	return GitServer{
		listenAddr: c.GetString("SERVER_ADDRESS"),
		gitUtil:    NewGitUtil(c),
	}
}
