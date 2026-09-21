---
title: "GCP 기본 환경 설정"
aliases:
  - "2장 GCP 환경"
chapter: "ch02"
type: "note"
status: "done"
created: 2026-09-19
updated: 2026-09-21
tags:
  - notiflex
  - ch02
  - gcp
  - environment
related:
  - "[[JOURNEY]]"
  - "[[02_github-repository]]"
---

# GCP 기본 환경 설정

## 연결 문서

- 상위 진행 기록: [[JOURNEY]]
- 다음 문서: [[02_github-repository]]
- 2장 마감: [[07_ch02-handoff]]

## 기본값

- 활성 계정: `timzero01@gmail.com`
- 기본 프로젝트: `tim-gitaiops-project`
- 기본 리전: `asia-northeast3`
- gcloud 설정 이름: `default`

## 설정 명령
```bash
gcloud config set project tim-gitaiops-project
gcloud config set compute/region asia-northeast3
```

## 확인 명령

```bash
gcloud auth list
gcloud config list
gcloud projects describe tim-gitaiops-project
gcloud compute regions describe asia-northeast3
gcloud auth configure-docker asia-northeast3-docker.pkg.dev --quiet
gcloud services list --enabled --filter='artifactregistry.googleapis.com' --project tim-gitaiops-project
gcloud artifacts locations list
```

## 현재 확인 결과

`gcloud config list` 기준으로 프로젝트와 리전 기본값은 설정되었다.

`gcloud projects describe tim-gitaiops-project` 실행 결과 프로젝트 접근도 정상 확인되었다.

- 프로젝트 상태: `ACTIVE`
- 프로젝트 번호: `782651363057`
- 생성 시간: `2026-09-19T11:28:52.757Z`

Compute Engine API 활성화 후 `gcloud compute regions describe asia-northeast3` 실행도 정상 확인되었다.

- 리전 이름: `asia-northeast3`
- 리전 상태: `UP`
- 사용 가능 존: `asia-northeast3-a`, `asia-northeast3-b`, `asia-northeast3-c`
- 주요 기본 할당량 예시: `CPUS=100`, `INSTANCES=24`, `DISKS_TOTAL_GB=2048`

이후 GKE 클러스터 생성 전에는 Kubernetes Engine API(`container.googleapis.com`) 활성화 여부를 확인해야 한다.

## Artifact Registry 인증

서울 리전 Artifact Registry Docker 도메인에 gcloud 인증 헬퍼를 설정했다.

```bash
gcloud auth configure-docker asia-northeast3-docker.pkg.dev --quiet
```

Docker 설정 파일 `~/.docker/config.json`에 다음 항목이 추가되었다.

```json
{
  "credHelpers": {
    "asia-northeast3-docker.pkg.dev": "gcloud"
  }
}
```

Artifact Registry API도 활성화했다.

```bash
gcloud services enable artifactregistry.googleapis.com --project tim-gitaiops-project
```

확인 결과 `artifactregistry.googleapis.com`이 활성화 목록에 있으며, Artifact Registry 위치 목록에 `asia-northeast3`가 표시된다.
