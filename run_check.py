import subprocess

res = subprocess.run(["go", "build", "./..."], cwd="D:\\agentic_devops\\agentic-cicd", capture_output=True, text=True)
print("STDOUT:", res.stdout)
print("STDERR:", res.stderr)
print("CODE:", res.returncode)
