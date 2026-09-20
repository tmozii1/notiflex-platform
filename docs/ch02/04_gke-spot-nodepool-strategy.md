# GKE Spot VM 노드풀 전략

## 결정

현재 생성된 GKE 클러스터 `notiflex-dev`는 그대로 유지한다.

다음에 클러스터를 재생성하거나 노드풀 구성을 다시 잡을 때는 노드를 다음 기준으로 만든다.

- 노드 수: `2`
- 노드 유형: Spot VM
- 머신 타입: `e2-medium`
- 위치: `asia-northeast3-a`

## 이유

Spot VM은 일반 VM보다 저렴하므로 학습용 GKE 클러스터 비용을 줄이는 데 유리하다.

다만 Spot VM은 Google Cloud가 언제든 회수할 수 있다. 노드가 1개뿐이면 Spot 회수 시 실습 앱이 바로 멈출 수 있으므로, 다음 재생성 시에는 Spot 노드를 2개로 구성해 최소한의 여유를 둔다.

## 현재 클러스터를 유지하는 이유

현재 클러스터는 정상 동작 중이다.

- 클러스터 이름: `notiflex-dev`
- 위치: `asia-northeast3-a`
- 노드 수: `1`
- 머신 타입: `e2-medium`
- 상태: `RUNNING`

당장 진행할 2장 실습에는 현재 구성으로도 문제가 없으므로, 지금은 불필요한 재생성 비용과 시간을 쓰지 않는다.

## 다음 재생성 시 사용할 명령 예시

클러스터를 새로 만들 때 Spot 노드 2개로 생성하는 예시:

```bash
gcloud container clusters create notiflex-dev \
  --zone asia-northeast3-a \
  --num-nodes 2 \
  --machine-type e2-medium \
  --disk-type pd-balanced \
  --disk-size 30 \
  --release-channel regular \
  --enable-ip-alias \
  --workload-pool tim-gitaiops-project.svc.id.goog \
  --spot
```

기존 클러스터를 유지한 상태에서 Spot 노드풀을 추가하는 예시:

```bash
gcloud container node-pools create spot-pool \
  --cluster notiflex-dev \
  --zone asia-northeast3-a \
  --num-nodes 2 \
  --machine-type e2-medium \
  --disk-type pd-balanced \
  --disk-size 30 \
  --spot
```

그 후 기본 노드풀을 줄이거나 삭제할 수 있다.

```bash
gcloud container node-pools resize default-pool \
  --cluster notiflex-dev \
  --zone asia-northeast3-a \
  --num-nodes 0
```

## 주의사항

- Spot VM은 언제든 회수될 수 있다.
- Stateful 워크로드나 장시간 배치에는 적합하지 않을 수 있다.
- 장애/무중단 배포 실습 전에는 Spot 회수로 인한 변수를 감안해야 한다.
- 운영 환경에서는 일반 노드풀과 Spot 노드풀을 섞는 구성이 더 안전하다.
