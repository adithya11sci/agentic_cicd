import os

with open(r'd:\agentic_devops\agentic-cicd\internal\orchestrator\orchestrator.go', 'r', encoding='utf-8') as f:
    orch = f.read()

orch = orch.replace('o.prAgent.CreatePR(ctx, event, repair)', 'o.prAgent.CreateFixPR(ctx, event, analysis, repair)')
with open(r'd:\agentic_devops\agentic-cicd\internal\orchestrator\orchestrator.go', 'w', encoding='utf-8') as f:
    f.write(orch)

with open(r'd:\agentic_devops\agentic-cicd\cmd\server.go', 'r', encoding='utf-8') as f:
    serv = f.read()

serv = serv.replace('prAgent := agents.NewPRAgent(logger, githubSvc)', 'prAgent := agents.NewPRAgent(githubSvc, logger)')
with open(r'd:\agentic_devops\agentic-cicd\cmd\server.go', 'w', encoding='utf-8') as f:
    f.write(serv)

with open(r'd:\agentic_devops\agentic-cicd\internal\agents\pragent.go', 'r', encoding='utf-8') as f:
    pr = f.read()

pr = pr.replace('analysis.Confidence', 'fmt.Sprintf("%.2f%%", analysis.Confidence*100)')
with open(r'd:\agentic_devops\agentic-cicd\internal\agents\pragent.go', 'w', encoding='utf-8') as f:
    f.write(pr)
