---
title: "2장 마감 Handoff"
aliases:
  - "2장 인수인계"
chapter: "ch02"
type: "handoff"
status: "done"
created: 2026-09-20
updated: 2026-09-21
tags:
  - notiflex
  - ch02
  - handoff
  - gitops-ready
related:
  - "[[JOURNEY]]"
  - "[[06_notiflex-app-implementation]]"
---

# 2장 마감 Handoff

## 연결 문서

- 전체 진행 기록: [[JOURNEY]]
- 직전 배포 기록: [[06_notiflex-app-implementation]]
- GitHub 공개 개요 문서: `README.md`

## 현재 위치

2장에서는 Notiflex API 서버를 만들고 GKE에 배포하는 첫 흐름을 완료했다.

완료된 범위:

- GCP 프로젝트 기본값 설정
- Artifact Registry 인증과 저장소 준비
- GKE Standard 클러스터 `notiflex-dev` 생성
- kubeconfig와 kubectl context 설정
- Go 기반 Notiflex API 서버 구현
- Docker 이미지 빌드와 Artifact Registry push
- Kubernetes Namespace, Deployment, Service 매니페스트 작성
- GKE 배포와 port-forward 기반 동작 확인
- 2장 문서를 `docs/ch02/`로 정리

## 현재 주요 값

```text
GCP 프로젝트: tim-gitaiops-project
리전: asia-northeast3
존: asia-northeast3-a
GKE 클러스터: notiflex-dev
Kubernetes 네임스페이스: notiflex
Artifact Registry: asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex
현재 앱 이미지: asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex/notiflex-api:v0.1.1
```

## 현재 배포 상태

마지막 확인 기준:

- `Deployment/notiflex-api` 롤아웃 완료
- Pod 1개 Running
- Service는 `ClusterIP`
- 외부 LoadBalancer는 만들지 않음
- 로컬 검증은 `kubectl port-forward -n notiflex svc/notiflex-api 18084:80`로 수행

## 2장에서 남긴 중요한 결정

- 현재 클러스터는 일반 `e2-medium` 1노드로 유지한다.
- 다음에 클러스터를 재생성할 때는 Spot VM 2노드 구성을 검토한다.
- 앱 컨테이너 최종 이미지는 현재 `alpine` 기반으로 유지한다.
- 운영 배포 단계에서 `scratch` 기반 이미지로 줄이는 것을 다시 검토한다.
- 현재 Artifact Registry 이미지는 Cloud Build가 아니라 로컬 Docker Buildx로 빌드해 push했다.
- Apple Silicon 환경에서는 GKE용 이미지를 `linux/amd64`로 명시해서 빌드해야 한다.

## 3장 시작 전 확인하면 좋은 명령

```bash
gcloud auth list
gcloud config list
kubectl config current-context
kubectl get nodes
kubectl get deploy,pods,svc -n notiflex
```

## 3장 시작 후보

3장은 ArgoCD 기반 GitOps 도입으로 이어간다.

추천 시작 흐름:

1. Obsidian 목차에서 3장 목표와 체크리스트를 확인한다.
2. 현재 클러스터와 배포 상태를 확인한다.
3. ArgoCD 설치 방식을 결정한다.
4. `deploy/argocd/` 구조를 만들고 Notiflex 앱을 GitOps 관리 대상으로 전환한다.
5. 수동 `kubectl apply` 흐름과 ArgoCD 동기화 흐름의 차이를 문서화한다.

## 다음 세션 참고 문서

- [GCP 기본 환경](01_gcp-environment.md)
- [GKE 클러스터](03_gke-cluster.md)
- [Spot VM 결정 기록](04_gke-spot-nodepool-strategy.md)
- [Notiflex 앱 설계](05_notiflex-app-design.md)
- [앱 구현과 배포 기록](06_notiflex-app-implementation.md)
- [루트 진행 기록](../../JOURNEY.md)
