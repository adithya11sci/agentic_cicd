import os

def replace_mod(path):
    with open(path, 'r', encoding='utf-8') as f:
        content = f.read()
    content = content.replace('"agentic-cicd/internal', '"github.com/user/agentic-cicd/internal')
    with open(path, 'w', encoding='utf-8') as f:
        f.write(content)

for root, _, files in os.walk(r'd:\agentic_devops\agentic-cicd'):
    for file in files:
        if file.endswith('.go'):
            replace_mod(os.path.join(root, file))
