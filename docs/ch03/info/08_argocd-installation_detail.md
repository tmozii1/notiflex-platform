---
title: "ArgoCD 설치 상세 개념"
aliases:
  - "3장 ArgoCD 설치 detail"
  - "ArgoCD Pod 역할"
chapter: "ch03"
type: "detail"
status: "done"
created: 2026-09-21
updated: 2026-09-21
tags:
  - notiflex
  - ch03
  - argocd
  - gitops
  - kubernetes
  - detail
related:
  - "[[08_argocd-installation]]"
  - "[[07_ch02-handoff]]"
  - "[[JOURNEY]]"
---

# ArgoCD 설치 상세 개념

## 연결 문서

- 실습 문서: [[08_argocd-installation]]
- 2장 마감 handoff: [[07_ch02-handoff]]
- 전체 진행 기록: [[JOURNEY]]

## 목적

이 문서는 3.2 ArgoCD 설치 실습 중 헷갈렸던 개념을 나중에 다시 보기 위해 정리한 보충 문서다.

실습을 끝까지 진행한 뒤 처음부터 다시 복습할 때, 여기의 용어와 구조를 다시 확인한다.

3장 마감 시점부터는 이 클러스터에서 직접 `kubectl apply`와 `kubectl delete`를 사용하지 않는다. 아래 설치 명령은 ArgoCD를 처음 도입할 때 실제로 사용한 역사적 기록이고, 이후 Kubernetes 리소스 변경은 Git 변경과 ArgoCD 동기화로 반영한다.

## 설치 명령의 의미

실행한 설치 명령은 다음과 같다.

```bash
kubectl create namespace argocd

kubectl apply -n argocd --server-side --force-conflicts \
  -f https://raw.githubusercontent.com/argoproj/argo-cd/v3.5.3/manifests/install.yaml
```

이 명령은 ArgoCD를 `kubectl`에 승인하는 명령이 아니다.

더 정확히는 다음을 수행한다.

```text
Kubernetes 클러스터 안에 ArgoCD를 설치한다.
ArgoCD가 동작하는 데 필요한 리소스와 권한을 만든다.
```

설치 매니페스트에는 여러 종류의 리소스가 들어 있다.

```text
CRD
ServiceAccount
Role / ClusterRole
RoleBinding / ClusterRoleBinding
ConfigMap
Secret
Service
Deployment
StatefulSet
NetworkPolicy
```

즉, 전체적으로는 다음 두 가지가 같이 일어난다.

```text
1. ArgoCD 프로그램 설치
2. ArgoCD가 Kubernetes API를 사용할 권한 부여
```

권한을 받는 주체는 `kubectl`이 아니라 클러스터 안에 설치되는 ArgoCD 컴포넌트다.

## `-n argocd`

`-n argocd`는 namespaced 리소스를 `argocd` 네임스페이스에 만들라는 뜻이다.

예를 들면 다음 리소스는 `argocd` 네임스페이스에 만들어진다.

```text
Deployment/argocd-server
Deployment/argocd-repo-server
Service/argocd-server
ConfigMap/argocd-cm
Secret/argocd-secret
```

반면 다음 리소스는 클러스터 전체 범위라 네임스페이스에 묶이지 않는다.

```text
CRD
ClusterRole
ClusterRoleBinding
```

## `-f https://.../install.yaml`

`-f`는 적용할 매니페스트 파일을 지정한다.

여기서는 로컬 파일이 아니라 GitHub에 있는 공식 설치 YAML을 URL로 지정했다.

```text
로컬 파일 적용: kubectl apply -f file.yaml
원격 파일 적용: kubectl apply -f https://.../install.yaml
```

이번 실습에서는 ArgoCD 공식 릴리스 `v3.5.3`의 설치 매니페스트를 사용했다.

## `--server-side`

`--server-side`는 Server-Side Apply를 사용하겠다는 뜻이다.

일반 apply는 `kubectl` 클라이언트가 변경 내용을 계산해 API 서버에 보낸다.

```text
kubectl 클라이언트가 이전 상태와 새 상태를 비교한다.
```

Server-Side Apply는 Kubernetes API 서버가 필드 단위로 변경과 소유권을 관리한다.

```text
Kubernetes API 서버가 어떤 manager가 어떤 필드를 관리하는지 기록한다.
```

장점:

- 큰 매니페스트와 CRD 적용에 더 적합하다.
- 필드 단위 소유권을 추적할 수 있다.
- 여러 컨트롤러가 같은 리소스를 다룰 때 충돌을 더 명확히 알 수 있다.
- 설치나 업그레이드 매니페스트에 안정적인 방식으로 쓰기 좋다.

ArgoCD 공식 Quick Start에서도 이 방식으로 설치한다.

## `--force-conflicts`

`--force-conflicts`는 Server-Side Apply에서 필드 소유권 충돌이 생길 때 이번 apply를 우선하겠다는 뜻이다.

예를 들어 어떤 필드를 기존의 다른 manager가 소유하고 있는데, 지금 적용하는 매니페스트도 같은 필드를 바꾸려 하면 충돌이 날 수 있다.

이 옵션을 붙이면 다음 의미가 된다.

```text
이번에 적용하는 매니페스트가 기준이다.
충돌나는 필드의 소유권을 이번 apply 쪽으로 가져온다.
```

따라서 이름처럼 충돌을 강제로 해결하는 옵션은 맞다.

다만 리소스를 무조건 삭제하는 파괴적 옵션은 아니다. Server-Side Apply의 필드 소유권 충돌을 강제로 넘기는 옵션이다.

운영 중 이미 커스터마이징된 ArgoCD에 다시 사용할 때는 조심해야 한다. 기존 설정의 일부 필드 소유권을 공식 매니페스트 쪽으로 가져올 수 있기 때문이다.

## ArgoCD Pod가 여러 개인 이유

ArgoCD는 단일 Pod 하나짜리 앱이 아니라 GitOps 운영 시스템이다.

그래서 설치하면 역할별 Pod가 여러 개 생성된다.

이번 실습에서 생성된 Pod:

```text
argocd-application-controller
argocd-applicationset-controller
argocd-dex-server
argocd-notifications-controller
argocd-redis
argocd-repo-server
argocd-server
```

전체 흐름은 다음처럼 볼 수 있다.

```text
사용자 / 브라우저 / CLI
        |
        v
argocd-server
        |
        +---- repo-server
        |
        +---- application-controller
        |
        +---- redis
        |
        +---- dex-server
        |
        +---- notifications-controller
        |
        +---- applicationset-controller
```

## `argocd-server`

ArgoCD의 Web UI와 API 서버다.

우리가 port-forward로 접속하는 대상이 이 컴포넌트다.

```bash
kubectl port-forward -n argocd svc/argocd-server 18080:443
```

역할:

- Web UI 제공
- ArgoCD API 제공
- 로그인 처리
- Application 목록과 상태 표시
- 사용자의 sync 요청 처리

쉽게 말하면 다음과 같다.

```text
사람이 ArgoCD와 대화하는 입구
```

## `argocd-application-controller`

GitOps의 핵심 컴포넌트다.

이 컨트롤러가 계속 다음 두 상태를 비교한다.

```text
Git에 선언된 desired state
클러스터의 실제 상태
```

우리가 만든 Application은 다음 Git 경로를 바라본다.

```text
repoURL: https://github.com/tmozii1/notiflex-platform.git
path: deploy/k8s
```

application-controller는 이 경로의 매니페스트와 실제 클러스터의 `notiflex` 리소스를 비교한다.

```text
같으면 Synced
다르면 OutOfSync
필요하면 Sync
```

쉽게 말하면 다음과 같다.

```text
GitOps 동기화 담당자
```

## `argocd-repo-server`

Git 저장소를 읽고 Kubernetes 매니페스트를 만들어주는 컴포넌트다.

ArgoCD는 단순 YAML뿐 아니라 Helm, Kustomize 같은 방식도 처리할 수 있다.

repo-server는 Git 저장소에서 소스를 가져와 최종 적용 대상 매니페스트를 계산한다.

Notiflex 현재 구조에서는 다음 일을 한다.

```text
GitHub 저장소 읽기
deploy/k8s 경로 읽기
Namespace / Deployment / Service YAML 반환
```

쉽게 말하면 다음과 같다.

```text
Git 저장소 해석 담당자
```

## `argocd-redis`

ArgoCD 내부 캐시 저장소다.

Git 저장소 정보, Application 상태, 계산 결과 등을 매번 새로 처리하면 느리기 때문에 Redis를 사용한다.

실습 중 직접 만질 일은 거의 없다.

쉽게 말하면 다음과 같다.

```text
ArgoCD 내부 성능 보조 저장소
```

## `argocd-dex-server`

Dex는 인증과 SSO를 담당하는 컴포넌트다.

나중에 GitHub, Google Workspace, LDAP 같은 외부 로그인과 연결할 때 사용한다.

예:

```text
GitHub 계정으로 ArgoCD 로그인
Google 계정으로 ArgoCD 로그인
팀별 RBAC 적용
```

지금은 기본 admin 로그인만 사용하므로 크게 체감되지는 않는다.

쉽게 말하면 다음과 같다.

```text
로그인 연동 담당자
```

## `argocd-notifications-controller`

Application 상태 변화를 외부로 알려주는 컴포넌트다.

예:

```text
Sync 성공
Sync 실패
Health Degraded
배포 완료
```

이런 이벤트를 Slack, email, webhook 등으로 보낼 수 있다.

지금은 알림을 설정하지 않았으므로 대기 중인 컴포넌트라고 보면 된다.

쉽게 말하면 다음과 같다.

```text
배포 상태 알림 담당자
```

## `argocd-applicationset-controller`

ApplicationSet은 여러 Application을 자동으로 만들어주는 기능이다.

지금은 `notiflex-api` Application 하나만 직접 만들었다.

하지만 나중에 다음 상황이 생기면 ApplicationSet이 유용하다.

```text
dev / staging / prod 환경별 앱 생성
여러 클러스터에 같은 앱 배포
고객사별 namespace에 앱 배포
여러 Git repo를 스캔해서 앱 자동 생성
```

Notiflex가 고객사별 환경을 가지게 되면 다음처럼 여러 Application을 만들 수 있다.

```text
notiflex-tenant-a
notiflex-tenant-b
notiflex-tenant-c
```

쉽게 말하면 다음과 같다.

```text
여러 Application 생성 자동화 담당자
```

## 지금 단계에서 꼭 기억할 것

현재 단계에서 핵심은 세 가지다.

```text
argocd-server
argocd-repo-server
argocd-application-controller
```

세 컴포넌트의 역할:

```text
argocd-repo-server가 GitHub 저장소의 deploy/k8s를 읽는다.
argocd-application-controller가 Git 상태와 클러스터 상태를 비교하고 동기화한다.
argocd-server가 결과를 UI와 API로 보여준다.
```

나머지는 운영 편의 기능이다.

```text
redis: 캐시
dex: 로그인/SSO
notifications: 알림
applicationset: 여러 앱 자동 생성
```

## Notiflex 현재 연결 구조

현재 구조는 다음과 같다.

```text
GitHub repo: tmozii1/notiflex-platform
  path: deploy/k8s
        |
        v
ArgoCD Application: notiflex-api
        |
        v
GKE namespace: notiflex
  - Namespace/notiflex
  - Deployment/notiflex-api
  - Service/notiflex-api
```

이제 Notiflex API의 원하는 상태는 Git의 `deploy/k8s` 경로다.

ArgoCD는 이 경로를 읽고 클러스터의 실제 상태와 비교한다.
