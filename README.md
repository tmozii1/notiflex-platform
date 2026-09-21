# Notiflex 인프라 구성/배포 실습

이 저장소는 책 **「AI 시대에 개발자가 알아야 할 인프라 구성 배포」**를 따라가며 진행하는 실습 프로젝트입니다.

실습의 중심 시나리오는 `Notiflex`입니다. Notiflex는 고객사의 서비스에서 발생하는 회원가입, 결제, 배송 같은 이벤트를 API로 받아 이메일, SMS, Push Alarm으로 발송하는 B2B SaaS 알림 플랫폼입니다.

이 프로젝트에서는 사용자가 Notiflex의 DevOps 엔지니어 역할을 맡고, Codex와 함께 GCP, GKE, GitOps, 관측 가능성, 무중단 배포, 멀티 테넌시, 이벤트 드리븐 운영을 단계적으로 실습합니다.

## 학습 목차와 진행 상황

책의 목차와 체크리스트는 Obsidian 문서에서 관리합니다.

```text
/Users/tim/project/timob/300_Book/310_AI/인프라 구성배포/목차.md
```

해당 문서는 학습 진행 상황, 장별 메모, Notiflex 시나리오 보강 내용을 기록하는 기준 문서입니다.

## 실습 로드맵

- 2장: GCP 환경 구성, GKE 클러스터 생성, 첫 Notiflex API 서버 배포
- 3장: ArgoCD를 통한 GitOps 배포 파이프라인 구성
- 4장: Prometheus, Grafana, Loki, Fluent Bit 기반 관측 가능성 구축
- 5장: Gateway API와 Blue/Green 배포를 통한 무중단 전환
- 6장: Valkey 캐시, Secret Manager, Canary 배포로 엔터프라이즈 기반 정비
- 7장: 멀티 노드풀, App of Apps, 네임스페이스 격리로 규모 확장
- 8장: Kafka, Tempo, CronJob으로 이벤트 드리븐 운영과 배치 자동화
- 9장: GitAIOps 운영 표준 정리

## 저장소 구조

초기에는 최소 구조로 시작하고, 실습이 진행되면서 필요한 폴더를 추가합니다.

```text
.
├── AGENTS.md
├── README.md
├── docs/
│   ├── ch02/
│   └── ch03/
├── app/
├── deploy/
│   ├── k8s/
│   └── argocd/
├── infra/
└── scripts/
```

## 문서 작성 규칙

프로젝트 진행 중 생성되는 문서는 `docs/` 폴더에 둡니다.

- 문서는 `docs/ch02/`, `docs/ch03/`처럼 장별 폴더에 정리합니다.
- `architecture`, `decisions`, `notes` 같은 유형별 폴더로 나누지 않습니다.
- 각 장에서 만든 설계, 결정 기록, 학습 노트, 운영 절차는 모두 해당 장 폴더에 둡니다.

문서 파일명은 생성 순서를 알 수 있도록 `01_`, `02_`, `03_` 형식의 번호를 앞에 붙입니다.

예시:

```text
docs/ch02/01_gcp-environment.md
docs/ch02/05_notiflex-app-design.md
docs/ch03/08_argocd-installation.md
docs/ch03/09_github-actions-ci.md
docs/ch03/10_ch03-handoff.md
```

비밀번호, 토큰, 서비스 계정 키, kubeconfig, `.env` 파일은 저장소에 커밋하지 않습니다.

## 현재 상태

- 2장 환경 구성과 첫 Notiflex API GKE 배포 완료
- 3장 ArgoCD 기반 GitOps 배포 전환 완료
- GitHub Actions로 테스트, 이미지 빌드, Artifact Registry push, 매니페스트 이미지 태그 갱신 자동화 완료
- ArgoCD가 Git 변경을 감지해 GKE에 자동 배포하는 흐름 검증 완료
- 현재 배포 이미지: `asia-northeast3-docker.pkg.dev/tim-gitaiops-project/notiflex/notiflex-api:git-186ebf3`

현재 역할 분리는 다음과 같습니다.

```text
GitHub Actions: 테스트, 이미지 빌드, 이미지 push, deployment.yaml 갱신
ArgoCD: Git 변경 감지, Kubernetes 배포, Sync/Health 감시
```

## 작업 방식

실습은 다음 흐름을 기본으로 진행합니다.

```text
탐색 -> 비교/판단 -> 실행 -> 검증 -> 기록
```

GCP나 Kubernetes처럼 비용 또는 운영 상태에 영향을 줄 수 있는 작업은 실행 전에 현재 계정, 프로젝트, 리전, 클러스터 컨텍스트를 확인합니다.
