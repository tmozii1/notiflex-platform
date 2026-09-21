---
title: "Notiflex 실습 여정"
aliases:
  - "Notiflex Journey"
chapter: "root"
type: "journey"
status: "active"
created: 2026-09-20
updated: 2026-09-21
tags:
  - notiflex
  - infra-practice
  - journey
related:
  - "[[01_gcp-environment]]"
  - "[[07_ch02-handoff]]"
  - "[[08_argocd-installation]]"
  - "[[09_github-actions-ci]]"
---

# Notiflex 실습 여정

이 문서는 Notiflex 인프라 구성/배포 실습을 진행하면서 실제로 수행한 작업, 사용한 명령어, 중간에 만난 문제와 해결 방법을 기록한다.

책의 세부 목차와 체크리스트는 Obsidian 문서를 기준으로 관리한다.

```text
/Users/tim/project/timob/300_Book/310_AI/인프라 구성배포/목차.md
```

프로젝트 저장소는 다음 위치에서 진행한다.

```text
/Users/tim/project/gcp
```

GitHub 원격 저장소는 다음과 같다.

```text
https://github.com/tmozii1/notiflex-platform
```

## 연결 문서

- 2장 시작 문서: [[01_gcp-environment]]
- 2장 마감 handoff: [[07_ch02-handoff]]
- 3장 ArgoCD 설치: [[08_argocd-installation]]
- 3장 GitHub Actions CI: [[09_github-actions-ci]]
- GitHub 공개 개요 문서: `README.md`

## 현재 상태 요약

- GCP 프로젝트: `tim-gitaiops-project`
- 기본 리전: `asia-northeast3`
- 기본 존: `asia-northeast3-a`
- Artifact Registry 호스트: `asia-northeast3-docker.pkg.dev`
- Artifact Registry 저장소: `notiflex`
- GKE 클러스터: `notiflex-dev`
- Kubernetes 네임스페이스: `notiflex`
- 앱 이미지: `asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex/notiflex-api:v0.1.1`
- 배포 방식: Kubernetes `Deployment` + `ClusterIP Service`
- 외부 노출 방식: 현재는 `kubectl port-forward`로만 검증

현재 Notiflex API는 GKE 클러스터 안에서 실행 중이며, 로컬에서 포트 포워딩으로 정상 응답을 확인했다.

```text
local curl
  -> kubectl port-forward
  -> Service notiflex-api
  -> Deployment / Pod
  -> Go API server
```

## 1.5 시나리오 정리

Notiflex는 B2B SaaS 알림 플랫폼이다.

고객사의 서비스에서 회원가입, 결제, 배송 같은 이벤트가 발생하면 고객사는 Notiflex API를 호출한다. Notiflex는 이벤트를 받아 이메일, SMS, Push Alarm 같은 발송 채널로 전달하는 역할을 맡는다.

이번 실습의 초점은 처음부터 완성된 SaaS를 만드는 것이 아니라, 작은 API 서버를 만들고 GKE에 배포한 뒤 GitOps, 관측 가능성, 무중단 배포, 시크릿 관리, 멀티테넌트 구조로 점진적으로 확장하는 것이다.

## 2.3 GCP 기본 환경 준비

로컬에서 `gcloud` CLI를 사용할 수 있도록 설치하고, 실습용 GCP 프로젝트와 기본 리전을 설정했다.

주요 작업:

- Homebrew로 Google Cloud CLI 설치
- `gcloud auth login`으로 로그인
- 기본 프로젝트를 `tim-gitaiops-project`로 설정
- 기본 리전을 서울 리전 `asia-northeast3`로 설정
- 기본 존을 `asia-northeast3-a`로 설정
- Artifact Registry Docker 인증 설정

사용한 주요 명령어:

```bash
brew install --cask google-cloud-sdk
gcloud auth login
gcloud config set project tim-gitaiops-project
gcloud config set compute/region asia-northeast3
gcloud config set compute/zone asia-northeast3-a
gcloud auth configure-docker asia-northeast3-docker.pkg.dev --quiet
```

확인에 사용한 명령어:

```bash
gcloud auth list
gcloud config list
gcloud compute regions describe asia-northeast3
```

학습한 점:

- 프로젝트 ID, 리전, Artifact Registry 주소는 보통 공개 저장소에 올라가도 민감정보는 아니다.
- 하지만 토큰, 서비스 계정 키, kubeconfig, `.env` 같은 인증 정보는 절대 커밋하면 안 된다.
- 클라우드 작업 전에는 항상 현재 계정과 프로젝트를 확인해야 한다.

## 2.4 GitHub 저장소 준비

실습용 GitHub 저장소를 만들고 로컬 저장소를 연결했다.

저장소:

```text
tmozii1/notiflex-platform
```

주요 작업:

- 로컬 git 초기화
- `.gitignore` 추가
- GitHub public 저장소 생성
- `origin` 원격 연결
- 초기 파일 push

사용한 주요 명령어:

```bash
git init
gh repo create tmozii1/notiflex-platform --public --source=. --remote=origin --push
```

이후 진행 원칙:

- `git commit`, `git push`는 사용자가 요청할 때만 진행한다.
- 실습 중간 산출물은 먼저 로컬에서 검증하고, 필요한 시점에만 커밋한다.

## 2.5 GKE 클러스터 준비

Notiflex API를 배포할 GKE Standard 클러스터를 생성했다.

현재 클러스터:

- 이름: `notiflex-dev`
- 위치: `asia-northeast3-a`
- 노드 수: 1
- 머신 타입: `e2-medium`
- 디스크: `pd-balanced`, 30GB
- Workload Identity 풀: `tim-gitaiops-project.svc.id.goog`

사용한 주요 명령어:

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

`kubectl`이 이 클러스터를 바라보도록 kubeconfig도 설정했다.

```bash
gcloud container clusters get-credentials notiflex-dev \
  --zone asia-northeast3-a \
  --project tim-gitaiops-project
```

확인에 사용한 명령어:

```bash
kubectl config current-context
kubectl get nodes
```

학습한 점:

- kubeconfig는 `kubectl`이 어떤 클러스터에 어떤 사용자로 접속할지 저장하는 설정 파일이다.
- context는 kubeconfig 안에서 클러스터, 사용자, 네임스페이스 조합을 가리키는 이름이다.
- 다음번 클러스터를 재생성할 때는 비용 절감을 위해 Spot VM 노드 2개 구성을 검토하기로 했다.
- 당장 실습에는 기존 1노드 일반 VM 구성을 유지한다.

## 2.6 Notiflex 앱 구현과 배포

### 앱 설계

처음 배포할 앱은 Go 표준 라이브러리 기반의 작은 HTTP API 서버로 설계했다.

설계 문서:

```text
docs/ch02/05_notiflex-app-design.md
```

초기 엔드포인트:

- `GET /`
- `GET /healthz`
- `GET /readyz`
- `POST /v1/events`

`POST /v1/events`는 `type`, `recipient`, `channel`을 검증한 뒤 이벤트를 수락하면 `202 Accepted`를 반환한다.

### 앱 구현

주요 파일:

```text
app/main.go
app/go.mod
app/Dockerfile
app/.dockerignore
```

구현한 내용:

- HTTP 서버
- JSON 응답 헬퍼
- 이벤트 입력 검증
- health/readiness 엔드포인트
- 구조화된 로그
- graceful shutdown
- `ReadHeaderTimeout` 설정

검증:

```bash
cd app
go test ./...
```

### Docker 이미지 빌드

Dockerfile은 멀티 스테이지 빌드로 작성했다.

현재 구조:

- builder stage: `golang:1.23-alpine`
- runtime stage: `alpine:3.20`
- non-root 사용자로 실행

책에서는 `scratch` 기반 이미지를 사용하지만, 현재 단계에서는 디버깅 편의성을 위해 `alpine`을 유지했다. 운영 배포 단계에서 디버깅이 끝나면 `scratch`로 줄이는 방향을 다시 검토한다.

학습한 점:

- 멀티 스테이지 빌드는 빌드 환경과 실행 환경을 분리하는 방식이다.
- `scratch`와 `alpine`은 최종 실행 이미지 선택의 문제다.
- 즉, 멀티 스테이지 빌드 안에서 최종 stage를 `alpine`으로 둘 수도 있고 `scratch`로 둘 수도 있다.

### Artifact Registry 이미지 저장소

Artifact Registry에 Docker 이미지 저장소를 만들었다.

```bash
gcloud artifacts repositories create notiflex \
  --repository-format=docker \
  --location=asia-northeast3 \
  --description="Notiflex container images" \
  --project=tim-gitaiops-project
```

### 이미지 빌드와 푸시

현재 Artifact Registry에 올라간 이미지는 Cloud Build가 아니라 로컬 Docker Buildx로 빌드해서 푸시한 이미지다.

처음에는 Apple Silicon 환경에서 기본 플랫폼으로 이미지가 만들어져 GKE에서 다음 문제가 발생했다.

```text
no match for platform in manifest: not found
```

GKE 노드가 `linux/amd64` 이미지를 기대하므로, 다음 명령으로 플랫폼을 명시해서 다시 빌드하고 푸시했다.

```bash
docker buildx build \
  --platform linux/amd64 \
  -t asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex/notiflex-api:v0.1.1 \
  --push ./app
```

책의 2.6.3 흐름에서는 Cloud Build로 빌드해서 Artifact Registry에 올리는 절차를 다룬다. 현재 실습은 로컬 buildx로 먼저 성공시킨 상태이므로, Cloud Build 방식은 다음 학습 단계에서 다시 실습 대상으로 남겨둔다.

## Kubernetes 매니페스트 작성

클러스터에 배포할 매니페스트 3개를 작성했다.

```text
deploy/k8s/namespace.yaml
deploy/k8s/deployment.yaml
deploy/k8s/service.yaml
```

각 파일의 역할:

- `Namespace`: Notiflex 리소스를 다른 실습 리소스와 분리한다.
- `Deployment`: Notiflex API Pod를 원하는 개수로 유지하고 롤링 업데이트를 담당한다.
- `Service`: Pod IP가 바뀌어도 안정적인 내부 접근 주소를 제공한다.

현재 Service는 `ClusterIP` 타입이다. 외부 Load Balancer 비용과 노출을 피하기 위해, 검증은 `kubectl port-forward`로 진행했다.

1노드 클러스터에서 롤링 업데이트 중 새 Pod가 뜰 공간이 부족해 `Insufficient cpu` 문제가 발생했다. 이를 피하기 위해 Deployment 전략을 다음처럼 조정했다.

```yaml
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 0
    maxUnavailable: 1
```

## 매니페스트 적용과 동작 확인

매니페스트를 클러스터에 적용했다.

```bash
kubectl apply -f deploy/k8s/
```

롤아웃 상태를 확인했다.

```bash
kubectl rollout status deployment/notiflex-api -n notiflex --timeout=120s
kubectl get deploy,rs,pods,svc -n notiflex -o wide
```

서비스를 로컬로 포트 포워딩해서 확인했다.

```bash
kubectl port-forward -n notiflex svc/notiflex-api 18084:80
```

검증 요청:

```bash
curl -i http://127.0.0.1:18084/
curl -i http://127.0.0.1:18084/healthz
curl -i http://127.0.0.1:18084/readyz
curl -i -X POST http://127.0.0.1:18084/v1/events \
  -H 'Content-Type: application/json' \
  -d '{"type":"signup","recipient":"user@example.com","channel":"email"}'
```

확인 결과:

- `/` 응답 정상
- `/healthz` 응답 정상
- `/readyz` 응답 정상
- `/v1/events` 요청은 `202 Accepted` 반환
- Pod 로그에서 이벤트 수락 로그 확인

## 현재까지의 주요 문서

프로젝트 문서는 `docs/` 아래에 장별 폴더를 만들고 생성 순서 번호를 붙여 관리한다.

```text
docs/ch02/01_gcp-environment.md
docs/ch02/02_github-repository.md
docs/ch02/03_gke-cluster.md
docs/ch02/04_gke-spot-nodepool-strategy.md
docs/ch02/05_notiflex-app-design.md
docs/ch02/06_notiflex-app-implementation.md
docs/ch02/07_ch02-handoff.md
docs/ch03/08_argocd-installation.md
docs/ch03/09_github-actions-ci.md
docs/ch03/info/08_argocd-installation_detail.md
```

## 3.2 ArgoCD 설치와 GitOps 연결

3장에서는 푸시 기반 수동 배포의 한계를 확인하고, ArgoCD로 GitOps 배포 흐름을 시작했다.

진행한 작업:

- ArgoCD `v3.5.3` Non-HA 매니페스트 설치
- `argocd` 네임스페이스 생성
- ArgoCD Pod 전체 Ready 확인
- `deploy/argocd/notiflex-application.yaml` 추가
- GitHub 저장소 `deploy/k8s` 경로를 Notiflex API의 desired state로 연결
- ArgoCD Application `notiflex-api`가 `Synced`, `Healthy` 상태임을 확인
- ArgoCD UI를 port-forward로 확인

현재 GitOps 기준:

```text
repoURL: https://github.com/tmozii1/notiflex-platform.git
targetRevision: main
path: deploy/k8s
destination namespace: notiflex
```

## 3.4 GitHub Actions CI 구성

GitHub Actions가 Notiflex API의 테스트, 이미지 빌드, Artifact Registry push, 매니페스트 이미지 태그 갱신을 담당하도록 워크플로를 추가했다.

생성한 파일:

```text
.github/workflows/notiflex-api-ci.yml
docs/ch03/09_github-actions-ci.md
```

구성한 인증:

```text
Workload Identity Pool: github-actions
Provider: github
Service Account: github-actions-notiflex@tim-gitaiops-project.iam.gserviceaccount.com
GitHub repository variables:
  - GCP_WORKLOAD_IDENTITY_PROVIDER
  - GCP_SERVICE_ACCOUNT
OIDC condition:
  - assertion.repository == 'tmozii1/notiflex-platform' && assertion.ref == 'refs/heads/main'
```

역할 분리:

```text
GitHub Actions: 테스트, 이미지 빌드, 이미지 push, deployment.yaml 갱신
ArgoCD: Git 변경 감지, Kubernetes 배포, Sync/Health 감시
```

## 중요한 학습 메모

### 공개 저장소에 올려도 되는 정보와 안 되는 정보

대체로 공개해도 되는 정보:

- GCP 프로젝트 ID
- 리전, 존
- Artifact Registry 주소
- Kubernetes Namespace 이름
- 일반적인 매니페스트 구조

절대 공개하면 안 되는 정보:

- 서비스 계정 키 JSON
- OAuth 토큰
- kubeconfig
- `.env`
- 비밀번호, API 키, webhook secret
- 개인 인증 캐시

### Cloud Build와 로컬 Docker 빌드

현재 완료된 빌드 방식은 로컬 Docker Buildx다.

Cloud Build는 GCP 관리 환경에서 이미지를 빌드하고 Artifact Registry로 푸시하는 방식이다. 팀 개발과 CI/CD로 넘어가면 Cloud Build 또는 GitHub Actions 같은 원격 빌드 방식을 쓰는 것이 자연스럽다.

### Alpine과 scratch

`alpine`은 아주 작은 Linux 배포판이라 셸과 기본 도구를 사용할 수 있어 초기 디버깅에 편하다.

`scratch`는 빈 이미지에 가까워 더 작고 공격 표면이 적지만, 셸이 없어 운영 중 디버깅은 어렵다.

현재 단계에서는 학습과 디버깅을 위해 `alpine`을 사용하고, 운영 배포 단계에서 `scratch` 전환을 다시 검토한다.

### GKE 1노드 실습 주의점

1노드 클러스터는 비용은 낮지만, 롤링 업데이트나 리소스 여유가 부족할 때 Pending Pod가 생기기 쉽다.

이번에는 `maxSurge: 0`, `maxUnavailable: 1`로 조정해 기존 Pod를 먼저 내리고 새 Pod를 띄우는 방식으로 해결했다.

## 다음 단계 후보

다음 실습에서 이어갈 수 있는 작업:

1. 책 2.6.3 흐름에 맞춰 Cloud Build 방식으로 이미지 빌드와 푸시를 다시 실습한다.
2. Obsidian 목차에서 완료한 항목을 체크하고 현재 진행 메모를 갱신한다.
3. 사용자가 요청하면 현재 변경 사항을 commit/push한다.
4. 3.3장에서 Git push만으로 Notiflex 배포가 이어지는지 확인한다.
5. GKE 클러스터를 재생성할 시점에 Spot VM 2노드 구성을 적용한다.

## 현재 작업 원칙

- 클라우드 리소스 변경 전 `gcloud auth list`, `gcloud config list`로 계정과 프로젝트를 확인한다.
- Kubernetes 변경 전 `kubectl config current-context`, `kubectl get nodes`로 대상 클러스터를 확인한다.
- 삭제, 재생성, 비용 증가 가능성이 있는 작업은 실행 전에 명령어를 먼저 설명한다.
- `git commit`, `git push`는 사용자가 명시적으로 요청할 때만 진행한다.
