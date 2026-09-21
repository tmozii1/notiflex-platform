---
title: "GitHub Actions CI와 ArgoCD 연결"
aliases:
  - "3장 GitHub Actions CI"
chapter: "ch03"
type: "note"
status: "done"
created: 2026-09-21
updated: 2026-09-21
tags:
  - notiflex
  - ch03
  - github-actions
  - ci
  - argocd
  - gitops
related:
  - "[[08_argocd-installation]]"
  - "[[08_argocd-installation_detail]]"
  - "[[JOURNEY]]"
---

# GitHub Actions CI와 ArgoCD 연결

## 연결 문서

- ArgoCD 설치: [[08_argocd-installation]]
- ArgoCD 상세 개념: [[08_argocd-installation_detail]]
- 전체 진행 기록: [[JOURNEY]]

## 목표

3.4장에서는 ArgoCD 도입 후에도 남아 있던 수동 작업을 GitHub Actions로 자동화한다.

ArgoCD는 Kubernetes 배포와 상태 감시를 담당한다. 하지만 ArgoCD는 앱 이미지를 빌드하지 않는다.

따라서 GitHub Actions는 다음 역할을 맡는다.

```text
Go 테스트
Docker 이미지 빌드
Artifact Registry push
deployment.yaml 이미지 태그 갱신 커밋
```

ArgoCD는 GitHub Actions가 갱신한 `deploy/k8s/deployment.yaml`을 감지하고 GKE에 배포한다.

## 생성한 워크플로

파일:

```text
.github/workflows/notiflex-api-ci.yml
```

워크플로 이름:

```text
Notiflex API CI
```

## 동작 흐름

Pull Request에서는 테스트만 수행한다.

```text
pull_request
  -> go test ./...
```

`main` 브랜치에 앱 코드가 push되면 빌드와 매니페스트 갱신까지 진행한다.

```text
push main
  -> go test ./...
  -> Docker image build
  -> Artifact Registry push
  -> deploy/k8s/deployment.yaml 이미지 태그 갱신
  -> GitHub Actions bot이 변경 커밋 push
  -> ArgoCD가 Git 변경 감지
  -> GKE 배포
```

## 이미지 태그 규칙

GitHub Actions는 커밋 SHA 앞 7자리를 사용해 이미지 태그를 만든다.

예:

```text
git-17e1ee3
```

최종 이미지 주소:

```text
asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex/notiflex-api:git-17e1ee3
```

`APP_VERSION` 값도 같은 태그로 갱신한다.

## GitHub Actions와 ArgoCD 역할 분리

GitHub Actions는 클러스터에 직접 배포하지 않는다.

```text
GitHub Actions:
  - 테스트
  - 이미지 빌드
  - 이미지 push
  - manifest 변경 커밋

ArgoCD:
  - Git 변경 감지
  - Kubernetes 배포
  - Sync / Health 상태 감시
  - drift 감지와 self-heal
```

이 구조를 유지하면 Kubernetes 배포 기준은 계속 Git이 된다.

워크플로 권한은 job별로 나누었다.

```text
test job:
  - contents: read

build-and-update-manifest job:
  - contents: write
  - id-token: write
```

PR 검증에는 저장소 읽기 권한만 사용하고, GCP 인증과 manifest 커밋 권한은 `main` 브랜치 push에서 실행되는 빌드 job에만 부여한다.

## GitHub Repository Variables

워크플로 실행을 위해 GitHub 저장소에 다음 variables를 등록했다.

```text
GCP_WORKLOAD_IDENTITY_PROVIDER
GCP_SERVICE_ACCOUNT
```

이 값들은 secret이 아니라 식별자에 가깝지만, GitHub Actions 설정값으로 관리한다.

현재 값:

```text
GCP_WORKLOAD_IDENTITY_PROVIDER=projects/782651363057/locations/global/workloadIdentityPools/github-actions/providers/github
GCP_SERVICE_ACCOUNT=github-actions-notiflex@tim-gitaiops-project.iam.gserviceaccount.com
```

## 필요한 GCP 권한

GitHub Actions가 Artifact Registry에 이미지를 push할 수 있도록 GCP Workload Identity Federation을 구성했다.

생성한 서비스 계정:

```text
github-actions-notiflex@tim-gitaiops-project.iam.gserviceaccount.com
```

Artifact Registry `notiflex` 저장소에는 다음 권한을 부여했다.

```text
roles/artifactregistry.writer
```

GitHub 저장소가 해당 서비스 계정을 사용할 수 있도록 다음 권한도 부여했다.

```text
roles/iam.workloadIdentityUser
```

허용된 GitHub 저장소:

```text
tmozii1/notiflex-platform
```

OIDC Provider에는 다음 조건을 걸었다.

```text
assertion.repository == 'tmozii1/notiflex-platform' && assertion.ref == 'refs/heads/main'
```

즉, 이 서비스 계정은 `tmozii1/notiflex-platform` 저장소의 `main` 브랜치에서 실행되는 GitHub Actions에만 빌려줄 수 있다.

## 지금 단계의 의미

이 워크플로가 도입되면 사람이 직접 하던 다음 명령이 자동화된다.

```bash
docker buildx build --platform linux/amd64 \
  -t asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex/notiflex-api:<tag> \
  --push ./app
```

그리고 매니페스트 이미지 태그 갱신도 자동화된다.

```yaml
image: asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex/notiflex-api:<tag>
```

GitHub Actions는 GKE 클러스터 권한을 갖지 않는다. 클러스터 배포는 ArgoCD가 Git 변경을 감지해 수행한다.

## 다음 작업

다음 단계에서는 앱 코드를 한 번 더 수정해 전체 자동화 흐름을 테스트한다.

```text
code push
  -> GitHub Actions build
  -> manifest commit
  -> ArgoCD sync
  -> GKE rollout
```
