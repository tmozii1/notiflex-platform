# GitHub 저장소 생성 절차

## 목표

- 로컬 실습 프로젝트를 git 저장소로 초기화한다.
- GitHub에 `tmozii1/notiflex-platform` 공개 저장소를 만든다.
- 현재 프로젝트 파일을 첫 커밋으로 올린다.

## 진행 순서와 명령어

### 1. GitHub CLI 인증 확인

```bash
gh auth status
```

GitHub CLI가 `tmozii1` 계정으로 로그인되어 있고, 저장소 생성에 필요한 `repo` 권한이 있는지 확인한다.

### 2. 로컬 git 저장소 초기화

```bash
git init
git branch -M main
```

현재 프로젝트 폴더를 git 저장소로 만들고 기본 브랜치를 `main`으로 맞춘다.

### 3. 커밋 대상 확인

```bash
git status --short
```

어떤 파일이 첫 커밋에 들어갈지 확인한다.

### 4. 첫 커밋 생성

```bash
git add .
git commit -m "chore: initialize notiflex platform practice"
```

README, AGENTS, 문서, `.gitignore`를 첫 커밋으로 기록한다.

### 5. GitHub 공개 저장소 생성과 푸시

```bash
gh repo create tmozii1/notiflex-platform --public --source=. --remote=origin --push
```

GitHub에 공개 저장소를 만들고, 현재 로컬 저장소를 `origin` 원격으로 연결한 뒤 `main` 브랜치를 푸시한다.

### 6. 원격 연결 확인

```bash
git remote -v
git status --short --branch
gh repo view tmozii1/notiflex-platform --json nameWithOwner,visibility,url,defaultBranchRef
```

원격 저장소 연결, 로컬 브랜치 상태, GitHub 저장소 접근을 확인한다.

## 생성 결과

- 저장소 URL: `https://github.com/tmozii1/notiflex-platform`
- 공개 여부: `PUBLIC`
- 기본 브랜치: `main`
- 원격 이름: `origin`
- 원격 주소: `git@github.com:tmozii1/notiflex-platform.git`

첫 커밋:

```text
c22f75b chore: initialize notiflex platform practice
```

첫 커밋에 포함된 파일:

- `.gitignore`
- `AGENTS.md`
- `README.md`
- `docs/notes/01_gcp-environment.md`
- `docs/notes/02_github-repository.md`

## 다음에 필요한 절차

1. 장별 실습을 진행할 때마다 의미 있는 단위로 커밋한다.
2. 새 프로젝트 문서는 `docs/` 아래에 만들고, 파일명은 다음 번호인 `03_`부터 사용한다.
3. GKE 실습 전에 Kubernetes Engine API(`container.googleapis.com`)를 활성화한다.
4. 컨테이너 이미지를 올리기 전에 Artifact Registry 저장소를 서울 리전에 생성한다.
5. 비용이 발생하는 GCP 리소스 생성 전에는 `gcloud config list`로 프로젝트와 리전을 확인한다.
