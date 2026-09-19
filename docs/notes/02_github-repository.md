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
gh repo view tmozii1/notiflex-platform --web
```

원격 저장소 연결, 로컬 브랜치 상태, GitHub 저장소 접근을 확인한다.
