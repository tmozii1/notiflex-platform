# GKE 클러스터 생성

## 목표

책 2.5장의 GKE 클러스터 생성, kubeconfig 설정, 클러스터 상태 확인을 진행한다.

## 선택한 구성

- 프로젝트: `tim-gitaiops-project`
- 리전: `asia-northeast3`
- 존: `asia-northeast3-a`
- 클러스터 이름: `notiflex-dev`
- 클러스터 유형: GKE Standard
- 노드 수: `1`
- 머신 타입: `e2-medium`
- 디스크: `pd-balanced`, `30GB`
- 릴리스 채널: `regular`
- 네트워크 모드: VPC-native(`--enable-ip-alias`)
- Workload Identity Pool: `tim-gitaiops-project.svc.id.goog`

Regional 클러스터는 여러 존에 노드가 생겨 비용이 커질 수 있으므로, 학습 초기 단계에서는 서울 리전의 단일 존 클러스터로 시작한다.

## 진행 순서와 명령어

### 1. 현재 gcloud 설정 확인

```bash
gcloud config list
```

확인한 값:

- 계정: `timzero01@gmail.com`
- 프로젝트: `tim-gitaiops-project`
- 기본 리전: `asia-northeast3`

### 2. Kubernetes Engine API 활성화

```bash
gcloud services enable container.googleapis.com --project tim-gitaiops-project
```

GKE 클러스터 생성을 위해 `container.googleapis.com` API를 활성화했다.

### 3. GKE 인증 플러그인 확인

```bash
gcloud components install gke-gcloud-auth-plugin --quiet
gcloud components list --filter='id:gke-gcloud-auth-plugin'
```

확인 결과 `gke-gcloud-auth-plugin`은 설치된 상태였다.

### 4. 기본 존 설정

```bash
gcloud config set compute/zone asia-northeast3-a
```

서울 리전의 `asia-northeast3-a` 존을 기본 존으로 설정했다.

### 5. 기존 클러스터 확인

```bash
gcloud container clusters list --project tim-gitaiops-project
```

생성 전 기존 클러스터는 없었다.

### 6. 클러스터 생성

처음에는 삭제 보호 비활성화 플래그를 포함해 실행했다.

```bash
gcloud container clusters create notiflex-dev \
  --zone asia-northeast3-a \
  --num-nodes 1 \
  --machine-type e2-medium \
  --disk-type pd-balanced \
  --disk-size 30 \
  --release-channel regular \
  --enable-ip-alias \
  --workload-pool tim-gitaiops-project.svc.id.goog \
  --no-enable-deletion-protection
```

현재 gcloud의 `container clusters create` 명령에서는 `--no-enable-deletion-protection` 플래그가 지원되지 않아 실패했다. 클러스터는 생성되지 않았다.

다음 명령으로 다시 실행했다.

```bash
gcloud container clusters create notiflex-dev \
  --zone asia-northeast3-a \
  --num-nodes 1 \
  --machine-type e2-medium \
  --disk-type pd-balanced \
  --disk-size 30 \
  --release-channel regular \
  --enable-ip-alias \
  --workload-pool tim-gitaiops-project.svc.id.goog
```

생성 결과:

- 클러스터 이름: `notiflex-dev`
- 위치: `asia-northeast3-a`
- 상태: `RUNNING`
- Kubernetes 버전: `1.35.7-gke.1222000`
- 머신 타입: `e2-medium`
- 노드 수: `1`
- Stack type: `IPV4`
- Master IP: `34.22.79.149`

생성 명령 완료 시 kubeconfig entry도 자동 생성되었다.

## kubeconfig 설정 확인

```bash
kubectl config current-context
```

확인 결과:

```text
gke_tim-gitaiops-project_asia-northeast3-a_notiflex-dev
```

## 클러스터 상태 확인

```bash
gcloud container clusters list --project tim-gitaiops-project
kubectl get nodes -o wide
kubectl get pods -A
```

노드 확인 결과:

- 노드 이름: `gke-notiflex-dev-default-pool-7ba0c1e5-1hpf`
- 노드 상태: `Ready`
- 노드 버전: `v1.35.7-gke.1222000`
- 내부 IP: `10.178.0.3`
- 외부 IP: `8.230.1.163`
- 컨테이너 런타임: `containerd://2.1.9`

클러스터 생성 직후 일부 시스템 파드는 `PodInitializing`, `ContainerCreating`, `Init` 상태였다. 잠시 후 다시 확인했을 때 시스템 파드는 대부분 `Running`으로 안정화되었고, `metrics-server`의 이전 파드 1개만 `Completed` 상태로 남아 있었다.

확인된 주요 네임스페이스:

- `kube-system`
- `gmp-system`
- `gke-managed-cim`

## 다음 단계

1. 필요한 경우 Artifact Registry 저장소를 서울 리전에 생성한다.
2. 2.6장에서 Notiflex Go API 서버를 만들고 컨테이너 이미지 빌드/푸시/배포를 진행한다.
