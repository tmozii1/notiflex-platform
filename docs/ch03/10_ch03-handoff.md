---
title: "3장 마감 handoff"
aliases:
  - "3장 handoff"
  - "GitOps 파이프라인 handoff"
chapter: "ch03"
type: "handoff"
status: "done"
created: 2026-09-21
updated: 2026-09-21
tags:
  - notiflex
  - ch03
  - handoff
  - argocd
  - github-actions
  - gitops
related:
  - "[[JOURNEY]]"
  - "[[08_argocd-installation]]"
  - "[[08_argocd-installation_detail]]"
  - "[[09_github-actions-ci]]"
---

# 3장 마감 handoff

## 연결 문서

- 전체 진행 기록: [[JOURNEY]]
- ArgoCD 설치: [[08_argocd-installation]]
- ArgoCD 상세 개념: [[08_argocd-installation_detail]]
- GitHub Actions CI: [[09_github-actions-ci]]

## 현재 상태

3장에서는 Notiflex API 배포 흐름을 수동 `kubectl apply` 중심에서 GitOps 중심으로 전환했다.

현재 배포 파이프라인:

```text
app 코드 변경
  -> main 브랜치 push
  -> GitHub Actions
  -> Go 테스트
  -> Docker 이미지 빌드
  -> Artifact Registry push
  -> deploy/k8s/deployment.yaml 이미지 태그 갱신 커밋
  -> ArgoCD 감지
  -> GKE rollout
```

현재 ArgoCD 상태:

```text
Application: notiflex-api
Sync Status: Synced
Health Status: Healthy
Revision: a38ce49751c54367770ab28a4de185452f712d05
```

현재 배포 이미지:

```text
asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex/notiflex-api:git-186ebf3
```

현재 앱 버전 응답:

```json
{"service":"notiflex-api","version":"git-186ebf3","environment":"dev","status":"running"}
```

## 3장에서 완료한 것

- 푸시 기반 수동 배포의 한계를 정리했다.
- ArgoCD를 `argocd` 네임스페이스에 설치했다.
- Notiflex API를 ArgoCD `Application`으로 등록했다.
- GitHub 저장소 `deploy/k8s` 경로를 Kubernetes desired state로 삼았다.
- 앱에 `/version` API를 추가하고 `v0.2.0` 배포를 확인했다.
- GitHub Actions로 테스트, 이미지 빌드, Artifact Registry push를 자동화했다.
- GitHub Actions가 `deployment.yaml` 이미지 태그를 갱신하고 커밋하도록 구성했다.
- ArgoCD가 GitHub Actions bot 커밋을 감지해 GKE에 자동 배포하는 것을 확인했다.
- 앞으로 이 클러스터에서는 직접 `kubectl apply`, `kubectl delete`를 사용하지 않는다는 운영 지침을 추가했다.

## 검증한 명령

GitHub Actions 상태:

```bash
gh run list --repo tmozii1/notiflex-platform --workflow "Notiflex API CI" --limit 3
```

ArgoCD 상태:

```bash
kubectl get application notiflex-api -n argocd -o wide
```

Deployment 이미지:

```bash
kubectl get deployment notiflex-api -n notiflex \
  -o jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
```

앱 버전 응답:

```bash
kubectl port-forward svc/notiflex-api -n notiflex 18081:80
curl -s http://127.0.0.1:18081/version
```

## 다음 장에서 기억할 점

4장은 관측 가능성 구축이다. Prometheus, Grafana, Loki, Fluent Bit, 알림 규칙을 다룰 예정이다.

중요한 운영 원칙:

- 이 클러스터에서는 직접 `kubectl delete`를 실행하지 않는다.
- Kubernetes 변경은 직접 `kubectl apply`하지 않고 Git 변경과 ArgoCD 동기화로 반영한다.
- 변경 전에는 항상 diff를 먼저 확인한다.
- ArgoCD 관리 리소스를 수동으로 바꾸면 self-heal에 의해 Git 상태로 되돌아갈 수 있다.
- GitHub Actions는 GKE 권한을 갖지 않는다. 이미지를 만들고 Git을 갱신하는 역할만 맡는다.
- Kubernetes 배포 권한은 ArgoCD가 담당한다.

## 다음 시작점

4장을 시작할 때 먼저 확인할 상태:

```bash
git status --short --branch
kubectl get application notiflex-api -n argocd -o wide
kubectl get deploy,pod,svc -n notiflex -o wide
kubectl get pods -n argocd
```

그다음 관측 가능성 도구를 GitOps로 설치할 수 있는 구조를 먼저 설계한다.
