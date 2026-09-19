# Notiflex 앱 빌드와 배포 설계

> **Summary**: GKE에 배포할 최소 Notiflex Go API 서버와 컨테이너/Kubernetes 배포 구조를 정의한다.
>
> **Author**: Codex
> **Created**: 2026-09-19
> **Status**: Draft
> **Level**: Starter

---

## 1. 배경

책 2.6장에서는 Notiflex 앱을 만들고 GKE에 배포한다. 현재 프로젝트에는 GCP 기본 설정, Artifact Registry 인증, GitHub 저장소, GKE 클러스터 `notiflex-dev`가 준비되어 있다.

이번 설계의 목적은 아직 복잡한 알림 시스템을 만들기보다, Notiflex의 첫 API 서버를 작게 만들고 GKE에 배포하는 전체 흐름을 학습하는 것이다.

## 2. 목표

- Go로 최소 API 서버를 만든다.
- Docker 이미지로 빌드한다.
- 서울 리전 Artifact Registry에 이미지를 푸시한다.
- GKE 클러스터 `notiflex-dev`에 배포한다.
- `kubectl port-forward`로 동작을 확인한다.
- 이후 ArgoCD, 관측 가능성, 무중단 배포 실습으로 확장할 수 있는 기본 구조를 만든다.

## 3. 범위

### 포함

- Go API 서버
- Dockerfile
- Kubernetes Namespace, Deployment, Service
- Artifact Registry 저장소 생성
- 로컬 실행 검증
- 컨테이너 빌드/푸시
- GKE 배포와 동작 확인

### 제외

- 실제 이메일/SMS/Push 발송
- 데이터베이스 저장
- 메시지 큐
- 인증/인가
- 외부 LoadBalancer 생성
- GitHub Actions 자동 빌드
- ArgoCD 배포 자동화

제외 항목은 이후 장에서 단계적으로 추가한다.

## 4. 시스템 개요

```text
사용자/테스트 요청
    |
    | curl / port-forward
    v
Kubernetes Service: notiflex-api
    |
    v
Deployment: notiflex-api
    |
    v
Pod: Go HTTP Server
```

초기 배포에서는 외부 IP를 만들지 않는다. 비용과 보안 노출을 줄이기 위해 `ClusterIP` Service를 만들고, 실습 확인은 `kubectl port-forward`로 한다.

## 5. 저장소 구조

구현 후 예상 구조:

```text
.
├── app/
│   ├── go.mod
│   ├── main.go
│   └── Dockerfile
├── deploy/
│   └── k8s/
│       ├── namespace.yaml
│       ├── deployment.yaml
│       └── service.yaml
└── docs/
    └── architecture/
        └── 05_notiflex-app-design.md
```

## 6. API 설계

### 6.1 `GET /`

서비스 기본 정보를 반환한다.

응답 예시:

```json
{
  "service": "notiflex-api",
  "version": "v0.1.0",
  "message": "Notiflex API is running"
}
```

### 6.2 `GET /healthz`

Kubernetes liveness probe용 엔드포인트다.

응답 예시:

```json
{
  "status": "ok"
}
```

### 6.3 `GET /readyz`

Kubernetes readiness probe용 엔드포인트다.

초기 버전에서는 외부 의존성이 없으므로 항상 `200 OK`를 반환한다.

응답 예시:

```json
{
  "status": "ready"
}
```

### 6.4 `POST /v1/events`

고객사의 서비스 이벤트를 수신하는 Notiflex 핵심 API의 첫 형태다.

요청 예시:

```json
{
  "type": "signup",
  "recipient": "demo@example.com",
  "channel": "email",
  "payload": {
    "userName": "Demo User"
  }
}
```

필드:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | yes | 이벤트 유형. 예: `signup`, `payment`, `delivery` |
| `recipient` | string | yes | 알림 수신자 |
| `channel` | string | no | `email`, `sms`, `push` 중 하나. 기본값은 `email` |
| `payload` | object | no | 이벤트별 추가 데이터 |

응답 예시:

```json
{
  "accepted": true,
  "eventId": "evt_20260919_000001",
  "message": "event accepted"
}
```

초기 버전에서는 이벤트를 저장하거나 발송하지 않는다. 서버 로그에 이벤트를 기록하고 `202 Accepted`를 반환한다.

## 7. 애플리케이션 설계

### 언어와 런타임

- Language: Go
- HTTP server: Go 표준 라이브러리 `net/http`
- Listen port: `8080`
- Config: 환경 변수

### 환경 변수

| Name | Default | Description |
|------|---------|-------------|
| `PORT` | `8080` | HTTP 서버 포트 |
| `APP_VERSION` | `v0.1.0` | 응답에 표시할 앱 버전 |

### 오류 처리

- 지원하지 않는 메서드는 `405 Method Not Allowed`
- 잘못된 JSON은 `400 Bad Request`
- 필수 필드 누락은 `400 Bad Request`
- 이벤트 수신 성공은 `202 Accepted`

## 8. 컨테이너 설계

Dockerfile은 멀티 스테이지 빌드를 사용한다.

- build stage: Go 이미지로 바이너리 빌드
- runtime stage: 작은 런타임 이미지로 실행
- container port: `8080`

이미지 이름:

```text
asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex/notiflex-api:v0.1.0
```

Artifact Registry 저장소:

```text
name: notiflex
location: asia-northeast3
format: docker
```

## 9. Kubernetes 설계

### Namespace

```text
notiflex
```

### Deployment

| Item | Value |
|------|-------|
| Name | `notiflex-api` |
| Namespace | `notiflex` |
| Replicas | `1` |
| Container port | `8080` |
| Image | `asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex/notiflex-api:v0.1.0` |
| Liveness probe | `GET /healthz` |
| Readiness probe | `GET /readyz` |

초기 클러스터 노드는 1개이므로 replica는 1개로 시작한다. 이후 무중단 배포 실습이나 Spot 2노드 재생성 이후 replica를 늘린다.

### Service

| Item | Value |
|------|-------|
| Name | `notiflex-api` |
| Namespace | `notiflex` |
| Type | `ClusterIP` |
| Port | `80` |
| Target Port | `8080` |

외부 LoadBalancer는 만들지 않는다. 확인은 다음 방식으로 한다.

```bash
kubectl port-forward -n notiflex svc/notiflex-api 8080:80
```

## 10. 빌드와 배포 순서

### 1. Artifact Registry 저장소 생성

```bash
gcloud artifacts repositories create notiflex \
  --repository-format=docker \
  --location=asia-northeast3 \
  --description="Notiflex container images"
```

이미 존재하면 생성하지 않고 기존 저장소를 사용한다.

### 2. 로컬 실행

```bash
cd app
go run .
```

확인:

```bash
curl http://localhost:8080/
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

### 3. Docker 이미지 빌드

```bash
docker build -t notiflex-api:v0.1.0 ./app
```

### 4. 이미지 태그와 푸시

```bash
docker tag notiflex-api:v0.1.0 \
  asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex/notiflex-api:v0.1.0

docker push \
  asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex/notiflex-api:v0.1.0
```

### 5. Kubernetes 배포

```bash
kubectl apply -f deploy/k8s/
```

### 6. 배포 상태 확인

```bash
kubectl get namespace notiflex
kubectl get deploy -n notiflex
kubectl get pods -n notiflex -o wide
kubectl get svc -n notiflex
```

### 7. 동작 확인

```bash
kubectl port-forward -n notiflex svc/notiflex-api 8080:80
```

다른 터미널에서:

```bash
curl http://localhost:8080/
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
curl -X POST http://localhost:8080/v1/events \
  -H 'Content-Type: application/json' \
  -d '{"type":"signup","recipient":"demo@example.com","channel":"email"}'
```

## 11. 검증 기준

- `go run .`으로 로컬 서버가 실행된다.
- `/`, `/healthz`, `/readyz`가 `200 OK`를 반환한다.
- `/v1/events`가 정상 요청에 `202 Accepted`를 반환한다.
- Docker 이미지가 로컬에서 빌드된다.
- 이미지가 Artifact Registry에 푸시된다.
- GKE에서 Pod가 `Running` 상태가 된다.
- GKE에서 Deployment가 `Available` 상태가 된다.
- `kubectl port-forward`를 통해 API가 응답한다.

## 12. 보안과 비용 고려

- 초기 버전에서는 외부 LoadBalancer를 만들지 않는다.
- Secret이나 API Key를 사용하지 않는다.
- `.env`, kubeconfig, 서비스 계정 키는 저장소에 커밋하지 않는다.
- Artifact Registry 주소와 프로젝트 ID는 공개되어도 되는 식별자지만, 인증 정보는 공개하지 않는다.
- 현재 GKE 클러스터는 비용이 발생하므로 장시간 미사용 시 정리 전략을 별도로 둔다.

## 13. 이후 확장 포인트

- 3장: GitHub Actions와 ArgoCD로 빌드/배포 자동화
- 4장: Prometheus 메트릭, Loki 로그, 알림 추가
- 5장: Gateway API와 Blue/Green 배포
- 6장: Valkey 캐시, Secret Manager, Canary 배포
- 8장: Kafka 이벤트 드리븐 구조

## 14. 검토 질문

구현 전에 확인하면 좋은 질문:

1. 초기 API 서버는 Go 표준 라이브러리만 사용할까, 아니면 Gin 같은 프레임워크를 사용할까?
2. `POST /v1/events` 응답의 `eventId`는 단순 생성값으로 충분한가?
3. 초기 Service는 `ClusterIP + port-forward`로 충분한가?
4. 컨테이너 이미지 태그는 `v0.1.0`으로 시작해도 괜찮은가?
5. Artifact Registry 저장소 이름은 `notiflex`로 확정할까?

## Related Documents

- [GCP 기본 환경 설정](../notes/01_gcp-environment.md)
- [GKE 클러스터 생성](../notes/03_gke-cluster.md)
- [GKE Spot VM 노드풀 전략](../decisions/04_gke-spot-nodepool-strategy.md)

## Version History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 0.1 | 2026-09-19 | 초기 설계 작성 | Codex |
