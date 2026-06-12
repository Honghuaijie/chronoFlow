# ChronoFlow UI

Admin console for ChronoFlow, designed for an internal single-team workflow. The stack is Vue 3, TypeScript, Ant Design Vue, Pinia, Vue Router, and Axios.

## Features

- Login.
- Executor list, create, edit, and delete.
- Job list, create, edit, delete, start, stop, manual run, and kill.
- Glue Shell editor.
- Job log list, filters, detail, Glue snapshot, and log content viewer.
- Polling for running logs.

## Local Startup

```bash
npm install
npm run dev
```

Default API proxy target:

```text
http://127.0.0.1:10003
```

Override when needed:

```bash
VITE_API_PROXY_TARGET=http://127.0.0.1:10003 npm run dev
```

## Build

```bash
npm run build
```

## Directory Conventions

```text
src/
├── api/          # HTTP requests and response unwrapping only.
├── stores/       # Pinia state, loading, pagination, and request orchestration.
├── views/        # Page containers.
├── components/   # Reusable components.
├── types/        # TypeScript types.
├── utils/        # Utilities.
├── router/       # Routes.
└── layouts/      # Admin layout.
```

Fixed call chain:

```text
view -> store -> api
```

Views must not call APIs directly or handle raw HTTP responses.

## Default Account

The account is configured by Admin. Default:

```text
admin / admin123
```

## Development Notes

- This is an operations console, not a marketing site.
- Prefer tables, filters, status tags, and clear confirmations for dangerous actions.
- If the same job is running, the manual run button should be disabled.
- Killing a job should move from `running` to `killing`, then to `killed` or `failed`.
- Protobuf JSON may return int64 fields as strings. Frontend IDs are handled as strings.
