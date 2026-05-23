package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"hackathon_DDD/state"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

func registerRoutes() {
	http.HandleFunc("/scale", scaleHandler)
	http.HandleFunc("/attack", attackHandler)
	http.HandleFunc("/now-target", nowTargetHandler)
	http.HandleFunc("/attack-history", attackHistoryHandler)
	http.HandleFunc("/restart", restartHandler)
}

// /scale へのリクエストを処理する関数
func scaleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "PUTで送ってください", http.StatusMethodNotAllowed)
		return
	}

	var req ScaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSONの形式が腐っています", http.StatusBadRequest)
		return
	}

	log.Printf("命令を受信。攻撃コンテナを %d 個にスケールします。\n", req.Count)

	// ここで docker compose up --scale attacker=N -d を実行
	cmd := exec.Command("docker", "compose", "up", "--scale", fmt.Sprintf("attacker=%d", req.Count), "-d")
	// 注意: docker-compose.yml があるディレクトリを指定します
	cmd.Dir = "./"

	if err := cmd.Run(); err != nil {
		log.Printf("スケール失敗: %v\n", err)
		http.Error(w, "Dockerの操作に失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "コンテナを %d 個に変更しましたよ\n", req.Count)
}

// /attack へのリクエストを処理する関数
func attackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POSTで送ってきなさい", http.StatusMethodNotAllowed)
		return
	}

	var req AttackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSONの形式が腐っています", http.StatusBadRequest)
		return
	}

	log.Printf("攻撃命令を受信。対象: %s, 回数: %d\n", req.Target, req.Count)

	// "all" の場合、稼働中のattackerのIPアドレスをDockerコマンドでかき集め
	if req.Target == "all" {
		// すべての attacker コンテナのIPアドレスを取得する
		cmd := exec.Command("docker", "ps", "-q", "--filter", "name=attacker")
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			http.Error(w, "コンテナの取得に失敗", http.StatusInternalServerError)
			return
		}

		containerIDs := strings.Fields(out.String())
		if len(containerIDs) == 0 {
			fmt.Fprintln(w, "attackerが1つも稼働していません")
			return
		}

		// docker exec で直接コンテナ内にコマンドを打ち込み
		for _, id := range containerIDs {
			go func(containerID string, count int) {
				// コンテナの中で、自分自身(127.0.0.1)のAPIをwgetコマンドで叩かせる
				targetUrl := fmt.Sprintf("http://127.0.0.1:80/fire?count=%d", count)
				execCmd := exec.Command("docker", "exec", containerID, "wget", "-qO-", targetUrl)

				if err := execCmd.Run(); err != nil {
					log.Printf("attacker(%s)への命令伝達に失敗: %v\n", containerID, err)
				}
			}(id, req.Count)

			log.Printf("attacker(%s)に %d 回の攻撃を指示しました\n", id[:8], req.Count)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%d 体のattackerに攻撃命令を下しました。\n", len(containerIDs))

	} else {
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintf(w, "指定コンテナへの個別命令はまだ未実装\n")
	}
}

// NowTargetへのリクエストを処理する関数
func nowTargetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GETで送ってください", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	data := getNowTarget()
	json.NewEncoder(w).Encode(data)
}

// AttckHistoryへのリクエストを処理する関数
func attackHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GETで送ってください", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	history := getAttackHistory()
	json.NewEncoder(w).Encode(history)
}

// restartTargetServer はtarget-server DeploymentのRolling Restartを発行する共通関数
func restartTargetServer() error {
	k8sClient, err := newK8sClient()
	if err != nil {
		return fmt.Errorf("Kubernetesクライアント初期化失敗: %w", err)
	}

	patch := `{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"` + metav1.Now().UTC().Format("2006-01-02T15:04:05Z") + `"}}}}}`
	_, err = k8sClient.AppsV1().Deployments("default").Patch(
		context.Background(),
		"target-server",
		types.StrategicMergePatchType,
		[]byte(patch),
		metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("Deploymentの再起動に失敗: %w", err)
	}

	log.Println("target-server Deploymentの再起動を発行しました")
	return nil
}

// /restart へのリクエストを処理する関数（フロントからの手動呼び出し用）
func restartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POSTで送ってください", http.StatusMethodNotAllowed)
		return
	}

	log.Println("リスタート命令を受信しました")

	if err := restartTargetServer(); err != nil {
		log.Printf("再起動失敗: %v\n", err)
		http.Error(w, "Podの再起動に失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RestartResponse{Message: "Podの再起動を開始しました"})
}

// Kubernetesクライアントを生成する（クラスター内→クラスター外の順にフォールバック）
func newK8sClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// ローカル開発時はkubeconfigを使用
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	return kubernetes.NewForConfig(config)
}

// Docker側で使うコンテナID取得関数（attack等で引き続き使用）
func getContainerID(containerName string) (string, error) {
	f := filters.NewArgs()
	f.Add("name", "^/"+containerName+"$")

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{
		Filters: f,
	})
	if err != nil {
		log.Printf("コンテナの取得に失敗しました: %v\n", err)
		return "", err
	}

	if len(containers) == 0 {
		return "", nil
	}

	return containers[0].ID, nil
}

// state更新はK8s移行後は不要だが、他で参照されている可能性があるためスタブとして残す
var _ = state.SetContainerID