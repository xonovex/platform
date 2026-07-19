# syntax=docker/dockerfile:1.7

FROM node:24.14.0-alpine3.23 AS dependencies
WORKDIR /dependencies
COPY packages/agent/agent-governance-decision-docker/runtime/package.json ./package.json
COPY packages/agent/agent-governance-decision-docker/runtime/package-lock.json ./package-lock.json
RUN npm ci --ignore-scripts --omit=dev

FROM node:24.14.0-alpine3.23 AS runtime
ENV NODE_ENV=production \
    DECISION_SERVICE_PORT=8787 \
    DECISION_EVIDENCE_PATH=/var/lib/xonovex/verdicts.jsonl
WORKDIR /app
COPY --from=dependencies --chown=node:node /dependencies/node_modules ./node_modules
COPY --chown=node:node packages/skill/skill-workflow/workflow-guide/scripts ./packages/skill/skill-workflow/workflow-guide/scripts
COPY --chown=node:node packages/skill/skill-agent-governance/agent-governance-guide/scripts ./packages/skill/skill-agent-governance/agent-governance-guide/scripts
RUN mkdir -p /var/lib/xonovex && chown node:node /var/lib/xonovex
VOLUME ["/var/lib/xonovex"]
EXPOSE 8787
USER node:node
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD node -e "fetch('http://127.0.0.1:8787/healthz').then(r=>{if(!r.ok)process.exit(1)}).catch(()=>process.exit(1))"
ENTRYPOINT ["node", "packages/skill/skill-agent-governance/agent-governance-guide/scripts/decision-service.ts"]
