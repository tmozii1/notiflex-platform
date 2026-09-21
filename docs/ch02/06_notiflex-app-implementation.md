---
title: "Notiflex 앱 구현과 GKE 배포"
aliases:
  - "2장 앱 구현"
chapter: "ch02"
type: "note"
status: "done"
created: 2026-09-19
updated: 2026-09-21
tags:
  - notiflex
  - ch02
  - app
  - deployment
  - gke
related:
  - "[[05_notiflex-app-design]]"
  - "[[07_ch02-handoff]]"
  - "[[JOURNEY]]"
---

# Notiflex 앱 구현과 GKE 배포

## 연결 문서

- 기준 설계: [[05_notiflex-app-design]]
- 2장 마감 handoff: [[07_ch02-handoff]]
- 전체 진행 기록: [[JOURNEY]]

## 목표

`docs/ch02/05_notiflex-app-design.md` 설계를 기준으로 Notiflex API 서버를 구현하고 GKE 클러스터 `notiflex-dev`에 배포한다.

## 생성한 파일

```text
app/
  go.mod
  main.go
  Dockerfile
  .dockerignore
deploy/
  k8s/
    namespace.yaml
    deployment.yaml
    service.yaml
```

## API 구현

Go 표준 라이브러리 `net/http`만 사용했다.

구현한 엔드포인트:

- `GET /`
- `GET /healthz`
- `GET /readyz`
- `POST /v1/events`

이벤트 수신 API는 실제 발송/저장 없이 요청을 검증하고 서버 로그에 기록한 뒤 `202 Accepted`를 반환한다.

## 로컬 검증

로컬에 Go가 없어 Homebrew로 Go CLI를 설치했다.

```bash
brew install go
go version
```

확인 결과:

```text
go version go1.27.1 darwin/arm64
```

포맷, 테스트, 빌드:

```bash
cd app
gofmt -w main.go
go test ./...
go build -o /tmp/notiflex-api .
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags=-s\ -w -o /tmp/notiflex-api-linux-amd64 .
```

검증 결과:

- `go test ./...`: 통과
- macOS 로컬 빌드: 통과
- linux/amd64 정적 바이너리 빌드: 통과

## 로컬 HTTP 검증

임시 포트 `18080`에서 실행했다.

```bash
PORT=18080 APP_VERSION=v0.1.0 /tmp/notiflex-api
```

확인:

```bash
curl -i http://localhost:18080/
curl -i http://localhost:18080/healthz
curl -i http://localhost:18080/readyz
curl -i -X POST http://localhost:18080/v1/events \
  -H 'Content-Type: application/json' \
  -d '{"type":"signup","recipient":"demo@example.com","channel":"email"}'
curl -i -X POST http://localhost:18080/v1/events \
  -H 'Content-Type: application/json' \
  -d '{"type":"signup"}'
```

결과:

- `/`: `200 OK`
- `/healthz`: `200 OK`
- `/readyz`: `200 OK`
- 정상 이벤트 요청: `202 Accepted`
- 필수 필드 누락: `400 Bad Request`

## Docker 이미지 빌드

초기 로컬 이미지 빌드:

```bash
docker build --progress=plain -t notiflex-api:v0.1.0 ./app
```

컨테이너 실행 검증:

```bash
docker run --rm -p 18081:8080 --name notiflex-api-test notiflex-api:v0.1.0
curl -i http://localhost:18081/healthz
curl -i -X POST http://localhost:18081/v1/events \
  -H 'Content-Type: application/json' \
  -d '{"type":"payment","recipient":"demo@example.com","channel":"push"}'
docker image inspect notiflex-api:v0.1.0 --format '{{.Config.User}} {{.Config.ExposedPorts}}'
```

확인 결과:

- 컨테이너 API 응답 정상
- 컨테이너 사용자: `notiflex`
- 노출 포트: `8080/tcp`

## Artifact Registry 저장소

저장소가 없어서 서울 리전에 생성했다.

```bash
gcloud artifacts repositories describe notiflex \
  --location=asia-northeast3 \
  --project=tim-gitaiops-project

gcloud artifacts repositories create notiflex \
  --repository-format=docker \
  --location=asia-northeast3 \
  --description="Notiflex container images" \
  --project=tim-gitaiops-project
```

저장소 URI:

```text
asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex
```

## 이미지 푸시 이슈와 해결

처음에는 Apple Silicon Docker 기본값으로 이미지를 빌드하고 `v0.1.0` 태그로 푸시했다.

```bash
docker tag notiflex-api:v0.1.0 \
  asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex/notiflex-api:v0.1.0

docker push \
  asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex/notiflex-api:v0.1.0
```

GKE 배포 후 Pod에서 다음 문제가 발생했다.

```text
ErrImagePull
no match for platform in manifest: not found
```

원인:

- 로컬 Mac은 Apple Silicon(`arm64`)
- GKE 노드는 `linux/amd64`
- 이미지 manifest가 GKE 노드 플랫폼과 맞지 않음

해결:

```bash
docker buildx build \
  --platform linux/amd64 \
  -t asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex/notiflex-api:v0.1.0 \
  --push ./app
```

이후 같은 문제를 피하기 위해 다음 수정 버전은 `v0.1.1` 태그로 빌드/푸시했다.

```bash
docker buildx build \
  --platform linux/amd64 \
  -t asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex/notiflex-api:v0.1.1 \
  --push ./app
```

## Kubernetes 배포

매니페스트 검증:

```bash
kubectl apply --dry-run=client -f deploy/k8s/
```

초기 적용:

```bash
kubectl apply -f deploy/k8s/
```

처음에는 파일 적용 순서 문제로 Namespace 생성 직후 Deployment가 실패했다.

```text
Error from server (NotFound): namespaces "notiflex" not found
```

해결:

```bash
kubectl apply -f deploy/k8s/deployment.yaml
```

## 1노드 클러스터 롤링 업데이트 이슈

`v0.1.1` 배포 중 새 Pod가 `Pending` 상태가 되었다.

확인 명령:

```bash
kubectl describe pod -n notiflex -l pod-template-hash=686f84cd8f
```

원인:

```text
0/1 nodes are available: 1 Insufficient cpu
```

현재 클러스터는 노드가 1개다. 기본 RollingUpdate 전략은 새 Pod를 먼저 띄운 뒤 기존 Pod를 내리려고 하므로, 순간적으로 Pod 2개가 필요해진다. GKE 시스템 Pod가 이미 자원을 사용 중이라 새 Pod 스케줄링이 실패했다.

해결:

Deployment 전략을 1노드 클러스터에 맞게 변경했다.

```yaml
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 0
    maxUnavailable: 1
```

적용:

```bash
kubectl apply -f deploy/k8s/deployment.yaml
kubectl rollout status deployment/notiflex-api -n notiflex --timeout=180s
```

결과:

```text
deployment "notiflex-api" successfully rolled out
```

## 최종 배포 상태

```bash
kubectl get pods -n notiflex -o wide
kubectl get deploy,svc -n notiflex
```

확인 결과:

- Pod: `1/1 Running`
- Deployment: `1/1 Available`
- Service: `ClusterIP`, `80 -> 8080`
- 이미지 태그: `v0.1.1`

## 최종 동작 확인

port-forward:

```bash
kubectl port-forward -n notiflex svc/notiflex-api 18083:80
```

확인:

```bash
curl -s http://localhost:18083/
curl -s -o /tmp/notiflex-health.out -w '%{http_code}' http://localhost:18083/healthz
curl -s -i -X POST http://localhost:18083/v1/events \
  -H 'Content-Type: application/json' \
  -d '{"type":"delivery","recipient":"demo@example.com","channel":"sms"}'
```

결과:

```json
{"message":"Notiflex API is running","service":"notiflex-api","version":"v0.1.1"}
```

```text
/healthz -> 200
POST /v1/events -> 202 Accepted
```

## 매니페스트 재적용 확인

2026-09-20에 2.6.4 흐름으로 Namespace, Deployment, Service 매니페스트를 다시 적용하고 동작을 확인했다.

적용 전 확인:

```bash
gcloud config list
kubectl config current-context
kubectl get nodes
kubectl apply --dry-run=client -f deploy/k8s/
```

실제 적용:

```bash
kubectl apply -f deploy/k8s/
kubectl rollout status deployment/notiflex-api -n notiflex --timeout=120s
kubectl get deploy,rs,pods,svc -n notiflex -o wide
```

결과:

- 매니페스트는 이미 적용된 상태라 `unchanged`로 확인되었다.
- Deployment rollout은 정상 완료되었다.
- Pod는 `1/1 Running` 상태였다.
- Service는 `ClusterIP`이며 `80 -> 8080`으로 연결된다.

동작 확인:

```bash
kubectl port-forward -n notiflex svc/notiflex-api 18084:80
curl -i http://localhost:18084/
curl -i http://localhost:18084/healthz
curl -i http://localhost:18084/readyz
curl -i -X POST http://localhost:18084/v1/events \
  -H 'Content-Type: application/json' \
  -d '{"type":"signup","recipient":"demo@example.com","channel":"email"}'
kubectl logs -n notiflex deploy/notiflex-api --tail=12
```

확인 결과:

- `/`: `200 OK`, 버전 `v0.1.1`
- `/healthz`: `200 OK`
- `/readyz`: `200 OK`
- `POST /v1/events`: `202 Accepted`
- Pod 로그에 이벤트 수신 로그가 기록됨

## 서브에이전트 리뷰 반영

서브에이전트가 구현 전/중 체크리스트를 제공했다. 반영한 주요 항목:

- `/healthz`와 `/readyz` 분리
- `POST /v1/events`의 `202 Accepted` 응답
- 잘못된 JSON과 필수 필드 누락의 `400 Bad Request`
- 지원하지 않는 메서드의 `405 Method Not Allowed`
- `ReadHeaderTimeout` 설정
- non-root 컨테이너 사용자
- `CGO_ENABLED=0` 정적 바이너리 빌드
- Service `80 -> 8080` 매핑
- Deployment/Service label selector 일치
- Artifact Registry 이미지 주소 검증
- public repo에 인증 정보 미포함

추가로 health/readiness probe 요청이 로그를 과도하게 남기지 않도록 `requestLogger`에서 `/healthz`, `/readyz`는 제외했다.

## 다음 단계

1. 2.7장의 `JOURNEY.md`를 생성한다.
2. Obsidian 목차에서 2.6 진행 상황을 업데이트한다.
3. 이후 3장에서 GitHub Actions와 ArgoCD를 도입해 수동 배포를 자동화한다.
