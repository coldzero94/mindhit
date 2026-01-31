# MindHit 제품 문서

> 브라우징 기록을 AI 마인드맵으로 변환하여, 흩어진 생각의 흐름을 한눈에 파악할 수 있는 서비스

## 문서 목록

| 순서 | 문서 | 설명 |
|------|------|------|
| 1 | [01-overview.md](./01-overview.md) | 제품 개요, 문제 정의, 솔루션, 타겟 사용자 |
| 2 | [02-user-flow.md](./02-user-flow.md) | 사용자 여정, 기능별 플로우 |
| 3 | [03-features.md](./03-features.md) | 핵심 기능 상세 명세 |
| 4 | [04-ui-wireframe.md](./04-ui-wireframe.md) | UI/UX 와이어프레임 |
| 5 | [05-pricing.md](./05-pricing.md) | 가격 정책, 플랜별 기능 비교 |

## 권장 읽기 순서

```
01-overview.md      ← 제품이 해결하는 문제와 핵심 가치
        ↓
02-user-flow.md     ← 사용자 관점의 서비스 흐름
        ↓
03-features.md      ← 각 기능의 상세 동작
        ↓
04-ui-wireframe.md  ← 화면 구성과 UI 설계
        ↓
05-pricing.md       ← 가격 정책과 플랜별 제한
```

## 핵심 개념

### 제품 구성
- **Chrome Extension**: 브라우징 기록 수집
- **Web App**: 세션 관리, 마인드맵 확인, 대시보드

### 플랜 구조
| 플랜 | 대상 | 특징 |
|------|------|------|
| Free | 체험용 | AI 토큰 50K/월, 세션 30일 보관 |
| Pro | 파워 유저 | AI 토큰 500K/월, 무제한 보관 |
| Enterprise | 팀/기업 | 무제한, 팀 협업 기능 |

## 개발 문서와의 관계

제품 문서는 **무엇을 만들 것인가**를 정의하고, [개발 문서](../development/README.md)는 **어떻게 만들 것인가**를 정의합니다.

```
docs/product/          → What to build (기획/디자인)
docs/development/      → How to build (개발/구현)
```
