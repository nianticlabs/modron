# Modron UI
User interface for the Modron service.

## Dependencies
- Docker 20.10 (and docker-compose)
- Node 16.16
- Angular CLI 14.1

## How to run

```bash
npm run genproto # Generate gRPC
npm run dev      # Run UI with mock gRPC server and envoy proxy
```

Then navigate to `localhost:4200`.