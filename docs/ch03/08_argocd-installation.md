---
title: "ArgoCD 설치와 GitOps 연결"
aliases:
  - "3장 ArgoCD 설치"
chapter: "ch03"
type: "note"
status: "done"
created: 2026-09-21
updated: 2026-09-21
tags:
  - notiflex
  - ch03
  - argocd
  - gitops
  - deployment
related:
  - "[[JOURNEY]]"
  - "[[07_ch02-handoff]]"
  - "[[06_notiflex-app-implementation]]"
  - "[[08_argocd-installation_detail]]"
  - "[[09_github-actions-ci]]"
  - "[[10_ch03-handoff]]"
---

# ArgoCD 설치와 GitOps 연결

## 연결 문서

- 전체 진행 기록: [[JOURNEY]]
- 2장 마감 handoff: [[07_ch02-handoff]]
- 직전 배포 기록: [[06_notiflex-app-implementation]]
- 설치 상세 개념: [[08_argocd-installation_detail]]
- 다음 CI 구성: [[09_github-actions-ci]]
- 3장 마감 handoff: [[10_ch03-handoff]]

## 목표

3.2장에서는 2장에서 수동으로 실행하던 `kubectl apply` 기반 배포를 GitOps 방식으로 전환하기 위한 첫 단계를 진행한다.

이번 단계의 목표는 다음과 같다.

- GKE 클러스터에 ArgoCD를 설치한다.
- ArgoCD가 GitHub 저장소의 Kubernetes 매니페스트를 바라보게 한다.
- Notiflex API를 ArgoCD `Application`으로 등록한다.
- Git에 선언된 상태와 클러스터의 실제 상태를 비교할 수 있게 만든다.

## 설치 전 확인

GCP와 Kubernetes 대상이 올바른지 먼저 확인했다.

```bash
gcloud auth list
gcloud config list
kubectl config current-context
kubectl get nodes
kubectl get deploy,pods,svc -n notiflex
```

확인한 값:

```text
계정: timzero01@gmail.com
프로젝트: tim-gitaiops-project
리전: asia-northeast3
존: asia-northeast3-a
kubectl context: gke_tim-gitaiops-project_asia-northeast3-a_notiflex-dev
GKE 노드 상태: Ready
Notiflex API 상태: Deployment 1/1, Pod Running, Service ClusterIP
```

## ArgoCD 설치

ArgoCD는 공식 릴리스 `v3.5.3`의 Non-HA 매니페스트로 설치했다.

학습용 1노드 클러스터이므로 HA 구성은 사용하지 않았다.

```bash
kubectl create namespace argocd
kubectl apply -n argocd --server-side --force-conflicts \
  -f https://raw.githubusercontent.com/argoproj/argo-cd/v3.5.3/manifests/install.yaml
```

설치 후 모든 ArgoCD Pod가 Ready 상태가 될 때까지 기다렸다.

```bash
kubectl wait --for=condition=Ready pod --all -n argocd --timeout=180s
kubectl get pods,deploy,sts,svc -n argocd -o wide
```

## GitOps 연결

Notiflex API를 ArgoCD가 관리하도록 `Application` 매니페스트를 추가했다.

파일:

```text
deploy/argocd/notiflex-application.yaml
```

핵심 설정:

```text
repoURL: https://github.com/tmozii1/notiflex-platform.git
targetRevision: main
path: deploy/k8s
destination namespace: notiflex
```

이 설정은 ArgoCD에게 다음 의미를 가진다.

```text
GitHub 저장소 main 브랜치의 deploy/k8s 경로가 Notiflex API의 원하는 상태다.
클러스터의 notiflex 네임스페이스 상태를 이 Git 상태와 계속 비교하고 동기화한다.
```

`syncPolicy.automated`에는 `prune`과 `selfHeal`을 켰다.

- `prune: true`: Git에서 제거된 리소스는 클러스터에서도 제거한다.
- `selfHeal: true`: 클러스터에서 수동 변경이 생기면 Git 상태로 되돌린다.

## 적용과 확인

ArgoCD Application을 클러스터에 적용했다.

```bash
kubectl apply -f deploy/argocd/notiflex-application.yaml
```

확인 명령:

```bash
kubectl get applications.argoproj.io -n argocd
kubectl describe application notiflex-api -n argocd
```

확인 결과:

```text
Application: notiflex-api
Sync Status: Synced
Health Status: Healthy
Revision: 126a2a579ad76fb05cf081da0af4d6dafa9994f2
Managed resources:
  - Namespace/notiflex
  - Service/notiflex-api
  - Deployment/notiflex-api
```

ArgoCD 자체 컴포넌트도 모두 Ready 상태를 확인했다.

```bash
kubectl get pods,deploy,sts,svc -n argocd
```

Notiflex API도 기존처럼 정상 Running 상태를 유지했다.

```bash
kubectl get deploy,pods,svc -n notiflex
```

UI는 외부 LoadBalancer를 만들지 않고 port-forward로 접근한다.

```bash
kubectl port-forward -n argocd svc/argocd-server 18080:443
```

port-forward 후 `https://127.0.0.1:18080`에서 HTTP 200 응답을 확인했다.

초기 admin 비밀번호는 다음 명령으로 확인할 수 있다.

```bash
kubectl get secret argocd-initial-admin-secret -n argocd \
  -o jsonpath='{.data.password}' | base64 -d
```

비밀번호는 문서나 git에 저장하지 않는다.

## 현재 의미

이제 Notiflex API는 ArgoCD가 Git 저장소의 `deploy/k8s` 경로를 기준으로 관리한다.

2장에서는 사람이 직접 다음 명령을 실행했다.

```bash
kubectl apply -f deploy/k8s/
```

3장부터는 이 기준이 바뀐다.

```text
GitHub 저장소의 deploy/k8s 경로
  -> ArgoCD Application
  -> GKE 클러스터의 notiflex 리소스
```

다음 실습에서는 Git 변경만으로 배포가 이어지는지 확인한다.
