<div align="center">
<pre>
███╗   ██╗ ██████╗ ████████╗██╗ ██████╗ ███╗   ██╗     ██████╗██╗     ██╗
████╗  ██║██╔═══██╗╚══██╔══╝██║██╔═══██╗████╗  ██║    ██╔════╝██║     ██║
██╔██╗ ██║██║   ██║   ██║   ██║██║   ██║██╔██╗ ██║    ██║     ██║     ██║
██║╚██╗██║██║   ██║   ██║   ██║██║   ██║██║╚██╗██║    ██║     ██║     ██║
██║ ╚████║╚██████╔╝   ██║   ██║╚██████╔╝██║ ╚████║    ╚██████╗███████╗██║
╚═╝  ╚═══╝ ╚═════╝    ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝     ╚═════╝╚══════╝╚═╝
</pre>

<p align="center">
  <a href="https://github.com/infograb/notion-cli/actions/workflows/ci.yml">
    <img src="https://github.com/infograb/notion-cli/actions/workflows/ci.yml/badge.svg" alt="CI/CD Pipeline">
  </a>
  <a href="https://codecov.io/gh/infograb/notion-cli">
    <img src="https://codecov.io/gh/infograb/notion-cli/branch/main/graph/badge.svg" alt="Code Coverage">
  </a>
  <a href="https://www.npmjs.com/package/@infograb/notion-cli">
    <img src="https://img.shields.io/npm/v/@infograb/notion-cli.svg" alt="npm version">
  </a>
  <a href="https://nodejs.org">
    <img src="https://img.shields.io/badge/node-%3E%3D22.0.0-brightgreen.svg" alt="Node.js Version">
  </a>
  <a href="https://github.com/infograb/notion-cli/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License">
  </a>
</p>
</div>

**중요 고지:**

이 프로젝트는 Notion API를 활용하는 독립적인 비공식 CLI 도구입니다.
Notion Labs, Inc.와 제휴, 보증, 후원 관계가 없습니다.
"Notion"은 Notion Labs, Inc.의 등록 상표입니다.

> AI 에이전트 & 자동화를 위한 Notion CLI (v5.9.0 · 5단계 성능 최적화)

## 프로젝트 소개

Notion API를 위한 강력한 CLI 도구로, AI 코딩 어시스턴트와 자동화 스크립트에 최적화되어 있습니다.

**주요 기능:**

- **빠른 네이티브 API 통합** — 인텔리전트 캐싱으로 최대 100배 빠른 반복 읽기
- **AI 우선 설계** — JSON 출력 모드, 구조화된 에러, 표준 종료 코드
- **비대화형** — 스크립트 및 자동화에 완벽히 적합
- **유연한 출력** — JSON, Markdown, Pretty, Compact JSON, Raw API 응답
- **최신 API** — Notion API v5.2.1, data sources 완전 지원
- **향상된 안정성** — 지수 백오프 자동 재시도 + 서킷 브레이커
- **5단계 성능 최적화** — 배치 작업 1.5-2배 성능 향상 (v5.9.0)
- **스키마 탐색** — AI 친화적 데이터베이스 스키마 추출
- **워크스페이스 캐싱** — API 호출 없이 빠른 데이터베이스 조회
- **스마트 ID 변환** — `database_id` ↔ `data_source_id` 자동 변환
- **보안** — 운영 취약점 0건

---

## 📦 설치 & 설정

```bash
# npm 설치 (권장)
npm install -g @infograb/notion-cli

# 최신 버전으로 업데이트
npm update -g @infograb/notion-cli
```

**최초 설정:**

```bash
# 대화형 설정 마법사 실행
notion-cli init
```

`init` 명령어가 다음을 안내합니다: API 토큰 설정 → 연결 테스트 → 워크스페이스 동기화

**수동 설정:**

```bash
export NOTION_TOKEN="secret_your_token_here"
notion-cli whoami   # 연결 확인
notion-cli sync     # 워크스페이스 캐시 동기화
```

토큰 발급: https://developers.notion.com/docs/create-a-notion-integration

---

## 🔑 핵심 기능

### Simple Properties — 복잡도 70% 감소

`-S` 플래그로 플랫 JSON을 사용해 페이지를 생성·수정합니다:

```bash
# 기존 방식 (중첩 구조)
notion-cli page create -d DB_ID --properties '{"Name": {"title": [{"text": {"content": "Task"}}]}, "Status": {"select": {"name": "Done"}}}'

# Simple Properties (-S 플래그)
notion-cli page create -d DB_ID -S --properties '{"Name": "Task", "Status": "Done", "Priority": 5, "Due Date": "tomorrow"}'
```

- 대소문자 구분 없음, 상대 날짜(`"today"`, `"+7 days"`) 지원
- 13가지 속성 타입: title, rich_text, number, checkbox, select, multi_select, status, date, url, email, phone, people, relation

### Smart ID Resolution — 자동 ID 변환

```bash
# database_id와 data_source_id 모두 사용 가능
notion-cli db retrieve 1fb79d4c71bb8032b722c82305b63a00  # database_id
notion-cli db retrieve 2gc80e5d82cc9043c833d93416c74b11  # data_source_id
```

### JSON Mode — AI 처리 최적화

```bash
notion-cli db query <ID> --json | jq '.data.results[].properties'
# 에러도 JSON 반환: {"success": false, "error": {"code": "NOT_FOUND", ...}}
```

종료 코드: `0` = 성공, `1` = API 오류, `2` = CLI 오류

### Schema Discovery — 스키마 탐색

```bash
notion-cli db schema <DATA_SOURCE_ID> --json            # 전체 스키마
notion-cli db schema <ID> --with-examples --json        # 복사 가능한 예시 포함
notion-cli db schema <ID> --properties Status,Priority  # 특정 속성만
```

### Workspace Caching — 워크스페이스 캐싱

```bash
notion-cli sync                           # 전체 워크스페이스 캐시 (1회)
notion-cli db query "Tasks Database"      # 이름으로 조회 (API 호출 없음)
notion-cli cache:info --json              # 캐시 상태 확인
```

---

## 📋 주요 명령어

### 설정 & 진단

| 명령어 | 설명 |
|--------|------|
| `notion-cli init` | 대화형 최초 설정 마법사 |
| `notion-cli doctor` | 7가지 헬스 체크 진단 |
| `notion-cli whoami` | 연결 및 봇 정보 확인 |
| `notion-cli config set-token` | API 토큰 설정 |

### 데이터베이스

```bash
notion-cli db retrieve <ID>                                # 메타데이터 조회
notion-cli db query <ID> --json                            # 쿼리
notion-cli db schema <ID> --json                           # 스키마 추출
notion-cli db update <ID> --title "New Title"              # 업데이트
notion-cli db create --parent-page <PAGE_ID> --title "DB"  # 생성
```

### 페이지

```bash
notion-cli page create -d <ID> -S --properties '{"Name": "Task"}'    # 생성
notion-cli page retrieve <PAGE_ID> --json                              # 조회
notion-cli page update <PAGE_ID> -S --properties '{"Status": "Done"}' # 수정
```

### 블록

```bash
notion-cli block retrieve <BLOCK_ID>                   # 블록 조회
notion-cli block append <BLOCK_ID> --children '[...]'  # 자식 추가
notion-cli block update <BLOCK_ID> --content "텍스트"  # 수정
```

### 사용자 & 검색

```bash
notion-cli user list --json               # 전체 사용자 목록
notion-cli user retrieve <USER_ID>        # 사용자 조회
notion-cli search "project" --json        # 워크스페이스 검색
notion-cli search "docs" --filter page    # 페이지로 필터링
```

### 워크스페이스

```bash
notion-cli sync          # 워크스페이스 동기화
notion-cli list --json   # 캐시된 데이터베이스 목록
```

---

## 🔍 데이터베이스 필터링

```bash
# 단일 조건
notion-cli db query <ID> --filter '{"property": "Status", "select": {"equals": "Done"}}' --json

# 복합 조건 (AND)
notion-cli db query <ID> \
  --filter '{"and": [{"property": "Status", "select": {"equals": "Done"}}, {"property": "Priority", "number": {"greater_than": 5}}]}' \
  --json

# 텍스트 검색
notion-cli db query <ID> --search "urgent" --json

# 파일에서 필터 로드
notion-cli db query <ID> --file-filter ./filter.json --json
```

자세한 내용: [docs/FILTER_GUIDE.md](./docs/FILTER_GUIDE.md)

---

## 📊 출력 형식

| 플래그 | 설명 |
|--------|------|
| `--json` | 구조화된 JSON (`{success, data, metadata}`) |
| `--compact-json` | 한 줄 압축 JSON |
| `--markdown` | Markdown 테이블 |
| `--pretty` | 테두리 있는 테이블 |
| `--raw` | 원시 API 응답 |

---

## ⚙️ 환경 변수

| 변수 | 기본값 | 설명 |
|------|--------|------|
| `NOTION_TOKEN` | — | Notion API 토큰 (필수) |
| `NOTION_RETRY_MAX_ATTEMPTS` | `3` | 최대 재시도 횟수 |
| `NOTION_RETRY_INITIAL_DELAY` | `1000` | 초기 지연 (ms) |
| `NOTION_RETRY_MAX_DELAY` | `30000` | 최대 지연 (ms) |
| `NOTION_CB_FAILURE_THRESHOLD` | `5` | 서킷 브레이커 임계값 |
| `NOTION_CACHE_DISABLED` | `false` | 전체 캐싱 비활성화 |
| `NOTION_CLI_VERBOSE` | `false` | stderr에 구조화 이벤트 로깅 |
| `NOTION_CLI_DEDUP_ENABLED` | `true` | 요청 중복 제거 |
| `NOTION_CLI_DELETE_CONCURRENCY` | `5` | 병렬 블록 삭제 수 |
| `NOTION_CLI_CHILDREN_CONCURRENCY` | `10` | 병렬 자식 페치 수 |
| `NOTION_CLI_DISK_CACHE_ENABLED` | `true` | 디스크 캐시 활성화 |
| `NOTION_CLI_DISK_CACHE_MAX_SIZE` | `104857600` | 최대 캐시 크기 (100MB) |
| `NOTION_CLI_HTTP_KEEP_ALIVE` | `true` | HTTP Keep-Alive |
| `DEBUG` | — | `notion-cli:*` 로 디버그 로깅 |

---

## 🚀 성능 최적화 (v5.9.0)

배치 작업과 반복 데이터 접근에서 **전체 1.5-2배 성능 향상**.

| 최적화 | 최대 효과 | 일반 효과 | 적용 시나리오 |
|--------|-----------|-----------|---------------|
| 요청 중복 제거 | 50% API 호출 감소 | 5-15% 감소 | 동시 중복 요청 |
| 병렬 작업 | 80% 빠름 | 60-70% 빠름 | 배치 삭제, 대량 조회 |
| 디스크 캐시 | 60% 히트율 향상 | 20-30% 향상 | 반복 CLI 세션 |
| HTTP Keep-Alive | 20% 빠름 | 5-10% 빠름 | 다중 요청 작업 |
| 응답 압축 | 70% 대역폭 절감 | 가변 | 대용량 JSON 응답 |

모든 최적화는 환경 변수로 독립 제어 가능하며, 하위 호환성 완전 유지.

---

## 🛠 개발 & 기여

**요구 사항:** Node.js >= 22.0.0, npm >= 8.0.0

```bash
# 저장소 클론 및 설치
git clone https://github.com/infograb/notion-cli
cd notion-cli && npm install && npm run build

# 개발 명령어
npm test              # 테스트 실행
npm run lint          # 린트 검사
npm run lint -- --fix # 자동 수정
npm link              # 전역 로컬 테스트
```

**기술 스택:** TypeScript · ESLint v9 · Prettier · Mocha + Chai

기여: Fork → 기능 브랜치 → 테스트 추가 → Pull Request

자세한 내용: [CONTRIBUTING.md](CONTRIBUTING.md) | [docs/](./docs/)

---

## 📄 라이선스 & 법적 고지

- **라이선스:** MIT — [LICENSE](LICENSE) 참고
- **상표:** "Notion"은 Notion Labs, Inc.의 등록 상표로, 이 프로젝트는 Notion Labs, Inc.와 무관한 독립 도구입니다.
- **서드파티:** [NOTICE](NOTICE) 파일에서 전체 오픈소스 라이선스 정보 확인

---

**문의:** [oss@infograb.net](mailto:oss@infograb.net) · [GitHub Issues](https://github.com/infograb/notion-cli/issues) · [npm](https://www.npmjs.com/package/@infograb/notion-cli)
