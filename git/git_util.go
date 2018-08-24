package git

import (
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/viper"
)

// GitUtil git 的工具类
type GitUtil struct {
	cachedRepoDir string
}

// FindLocalRepoPath 查找仓库对应的本地缓存仓库路径
func (g GitUtil) FindLocalRepoPath(remoteRepoURL string) string {
	localRepoRelativePath := strings.Replace(remoteRepoURL, "http://", "", -1)
	localRepoRelativePath = strings.Replace(localRepoRelativePath, "https://", "", -1)
	return path.Join(g.cachedRepoDir, localRepoRelativePath)
}

// CloneIfNotExist 如果 repo 不存在，从远程仓库克隆下来
func (g GitUtil) CloneIfNotExist(remoteRepoURL string) error {
	localRepoPath := g.FindLocalRepoPath(remoteRepoURL)
	_, err := os.Stat(localRepoPath)
	if !os.IsNotExist(err) {
		return nil
	}
	log.Println("创建目录" + localRepoPath)
	err = os.MkdirAll(localRepoPath, os.ModePerm)
	if err != nil {
		return err
	}
	cloneCmd := GitCommand{
		Args: []string{"git", "clone", "--mirror", remoteRepoURL, localRepoPath},
	}
	_, err = cloneCmd.Run(true)
	if err != nil {
		return err
	}

	setRemoteCmd := GitCommand{
		Args: []string{"git", "-C", localRepoPath, "remote", "set-url", "origin", remoteRepoURL},
	}
	_, err = setRemoteCmd.Run(true)
	return err
}

// func (g GitUtil) GitReceivePack(remoteRepoURL string, input *bytes.Reader) GitCommand {
// 	localRepoPath := g.FindLocalRepoPath(remoteRepoURL)
// 	cmd := GitCommand{
// 		ProcInput: input,
// 		Args:      []string{"receive-pack", "--stateless-rpc", localRepoPath}}
// 	return cmd
// }

// func (g GitUtil) Clone(remoteRepoURL string) error {
// 	localRepoPath := g.FindLocalRepoPath(remoteRepoURL)
// 	fmt.Println(localRepoPath)
// 	_, err := os.Stat(localRepoPath)
// 	if os.IsNotExist(err) {
// 		fmt.Println("创建目录" + localRepoPath)
// 		err := os.MkdirAll(localRepoPath, os.ModePerm)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 	} else {
// 		fmt.Println("目录已经存在")
// 		return nil
// 	}
// 	fmt.Println("准备 clone")
// 	cmd := exec.Command("git", "clone", "--mirror", remoteRepoURL, localRepoPath)
// 	return cmd.Run()
// }

func (g GitUtil) Fetch(localRepoPath string) error {
	cmd := exec.Command("git", "-C", localRepoPath, "fetch", "--quiet")
	return cmd.Run()
}

func NewGitUtil(c *viper.Viper) GitUtil {
	return GitUtil{
		cachedRepoDir: c.GetString("CACHED_REPO_DIR"),
	}
}
