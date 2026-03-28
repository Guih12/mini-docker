# 🐳 Mini-Docker em Go — Roadmap de Estudos e Implementação

> **Objetivo:** Construir um container runtime simplificado em Go, entendendo por dentro como o Docker funciona.
>
> **Linguagem:** Go
> **Nível:** Intermediário em Linux/Sistemas
> **Plataforma:** Linux (obrigatório — as syscalls são específicas do kernel Linux)

---

## Como usar este roadmap

Cada **fase** é uma etapa incremental do projeto. Dentro de cada fase, existem **issues** numeradas que representam tarefas concretas de implementação. A ideia é que você crie essas issues no seu repositório (GitHub/GitLab) e vá fechando conforme avança.

Marque com `[x]` conforme for completando:

---

## Fase 0 — Setup do Projeto e Fundamentos

Antes de escrever o container runtime, prepare o ambiente e estude os conceitos base.

- [x] **#001 — Inicializar repositório Go**
  - Criar o repo com `go mod init github.com/seuuser/mini-docker`
  - Estrutura sugerida: `cmd/`, `pkg/`, `internal/`, `examples/`
  - Configurar `.gitignore`, `Makefile`, `README.md`

- [ ] **#002 — Estudar conceitos fundamentais de containers**
  - Entender que containers NÃO são VMs
  - Ler sobre: namespaces, cgroups, chroot, union filesystems
  - Recurso: `man namespaces`, `man cgroups`, artigo "Containers from Scratch" (Liz Rice)

- [ ] **#003 — Estudar syscalls relevantes em Go**
  - Explorar o pacote `syscall` e `golang.org/x/sys/unix`
  - Praticar: `syscall.Exec`, `syscall.Sethostname`, `syscall.Chroot`
  - Entender `os/exec.Cmd` e os campos `SysProcAttr`

- [ ] **#004 — Criar o CLI básico com Cobra**
  - `go install github.com/spf13/cobra-cli@latest`
  - Implementar os comandos: `mini-docker run <cmd>`, `mini-docker ps`, `mini-docker exec`
  - Nesta etapa, `run` apenas executa o comando normalmente (sem isolamento)

---

## Fase 1 — Isolamento com Linux Namespaces

O coração de um container é o isolamento de processos via namespaces.

- [ ] **#005 — Implementar UTS namespace (hostname)**
  - Usar `CLONE_NEWUTS` no `SysProcAttr`
  - O processo filho deve ter seu próprio hostname
  - Testar: dentro do container, `hostname` deve ser diferente do host

- [ ] **#006 — Implementar PID namespace**
  - Usar `CLONE_NEWPID`
  - O processo dentro do container deve se ver como PID 1
  - Montar `/proc` dentro do container para que `ps` funcione

- [ ] **#007 — Implementar Mount namespace**
  - Usar `CLONE_NEWNS`
  - O container deve ter sua própria árvore de mounts
  - Implementar `mount("proc", "/proc", "proc", 0, "")` dentro do namespace

- [ ] **#008 — Implementar Network namespace (básico)**
  - Usar `CLONE_NEWNET`
  - Nesta fase, o container fica sem rede (isolado)
  - Verificar com `ip link` dentro do container

- [ ] **#009 — Implementar IPC namespace**
  - Usar `CLONE_NEWIPC`
  - Isolar filas de mensagens e memória compartilhada

- [ ] **#010 — Implementar User namespace (opcional, avançado)**
  - Usar `CLONE_NEWUSER`
  - Mapear UID/GID para rodar containers sem root
  - Configurar `/proc/PID/uid_map` e `/proc/PID/gid_map`

- [ ] **#011 — Técnica /proc/self/exe**
  - Implementar o padrão de re-invocar o próprio binário
  - O binário detecta via `os.Args` se é o processo "pai" ou "filho"
  - Isso permite configurar namespaces antes de exec

---

## Fase 2 — Sistema de Arquivos (Filesystem)

O container precisa de um filesystem root isolado.

- [ ] **#012 — Implementar chroot básico**
  - Baixar um rootfs mínimo (Alpine mini rootfs tar.gz)
  - Extrair em um diretório e fazer `syscall.Chroot()`
  - Testar: `ls /` dentro do container deve mostrar o rootfs isolado

- [ ] **#013 — Migrar de chroot para pivot_root**
  - `pivot_root` é mais seguro que `chroot`
  - Criar um mount point, chamar `syscall.PivotRoot()`, desmontar o antigo root
  - Lidar com o requisito de que new_root deve ser um mount point

- [ ] **#014 — Montar filesystems essenciais**
  - Montar `/proc` (procfs)
  - Montar `/sys` (sysfs)
  - Montar `/dev` (devtmpfs ou criar device nodes manualmente)
  - Montar `/dev/pts` (para pseudo-terminais)
  - Montar `/tmp` (tmpfs)

- [ ] **#015 — Implementar sistema de imagens simples**
  - Criar um diretório local `images/` com rootfs pré-extraídos
  - Comando: `mini-docker pull alpine` (baixa e extrai o rootfs)
  - Manter um registro simples em JSON: `{name, path, created_at}`

- [ ] **#016 — Implementar Copy-on-Write com OverlayFS**
  - Criar camadas: `lowerdir` (imagem, read-only), `upperdir` (mudanças), `workdir`, `merged`
  - Montar com: `mount -t overlay overlay -o lowerdir=...,upperdir=...,workdir=... merged/`
  - Cada container tem seu próprio `upperdir`, mas compartilha a imagem base

---

## Fase 3 — Controle de Recursos com cgroups

Limitar CPU, memória e processos do container.

- [ ] **#017 — Entender a hierarquia de cgroups v2**
  - Estudar `/sys/fs/cgroup/`
  - Entender controllers: `cpu`, `memory`, `pids`
  - Saber a diferença entre cgroups v1 e v2

- [ ] **#018 — Implementar limite de memória**
  - Criar cgroup: `mkdir /sys/fs/cgroup/mini-docker/<container-id>`
  - Escrever limite em `memory.max` (ex: "100M")
  - Adicionar PID do container em `cgroup.procs`
  - Testar: rodar um programa que aloca mais memória que o limite

- [ ] **#019 — Implementar limite de CPU**
  - Escrever em `cpu.max` (ex: "50000 100000" = 50% de 1 core)
  - Testar: rodar um processo CPU-bound e verificar uso

- [ ] **#020 — Implementar limite de processos (PIDs)**
  - Escrever em `pids.max` (ex: "20")
  - Testar: fork bomb deve ser contida

- [ ] **#021 — Cleanup de cgroups**
  - Ao parar o container, remover o diretório do cgroup
  - Implementar cleanup gracioso com `defer` e signal handling

---

## Fase 4 — Networking

Dar conectividade de rede ao container.

- [ ] **#022 — Criar bridge network no host**
  - Criar uma bridge: `ip link add mini-docker0 type bridge`
  - Atribuir IP: `ip addr add 172.20.0.1/24 dev mini-docker0`
  - Subir a interface: `ip link set mini-docker0 up`

- [ ] **#023 — Criar veth pair para o container**
  - Criar par: `ip link add veth-host type veth peer name veth-container`
  - Conectar `veth-host` à bridge
  - Mover `veth-container` para o network namespace do container
  - Configurar IP dentro do container: `ip addr add 172.20.0.2/24 dev veth-container`

- [ ] **#024 — Implementar NAT para acesso à internet**
  - Habilitar IP forwarding: `echo 1 > /proc/sys/net/ipv4/ip_forward`
  - Regra iptables: `iptables -t nat -A POSTROUTING -s 172.20.0.0/24 -j MASQUERADE`
  - Testar: `ping 8.8.8.8` de dentro do container

- [ ] **#025 — Implementar port mapping básico**
  - Flag: `mini-docker run -p 8080:80 alpine`
  - Regra iptables: `iptables -t nat -A PREROUTING -p tcp --dport 8080 -j DNAT --to 172.20.0.2:80`
  - Testar: acessar um servidor web rodando no container via host

---

## Fase 5 — Lifecycle de Containers

Gerenciar o ciclo de vida dos containers.

- [ ] **#026 — Gerar container IDs**
  - Gerar IDs aleatórios de 12 caracteres hex (como o Docker)
  - Manter metadados em `/var/lib/mini-docker/containers/<id>/config.json`

- [ ] **#027 — Implementar `mini-docker ps`**
  - Listar containers em execução
  - Mostrar: ID, comando, status, PID, data de criação
  - Ler informações de `/proc/<pid>` e do config.json

- [ ] **#028 — Implementar `mini-docker stop`**
  - Enviar `SIGTERM` para o PID 1 do container
  - Timeout de 10 segundos, depois `SIGKILL`
  - Limpar cgroups e network

- [ ] **#029 — Implementar `mini-docker rm`**
  - Remover diretório do container (upperdir do overlayfs)
  - Remover metadados
  - Verificar que o container está parado antes de remover

- [ ] **#030 — Implementar `mini-docker exec`**
  - Entrar nos namespaces de um container em execução via `setns()`
  - Usar `/proc/<pid>/ns/*` para obter os file descriptors dos namespaces
  - Executar um comando dentro do container existente

- [ ] **#031 — Implementar `mini-docker logs`**
  - Redirecionar stdout/stderr do container para um arquivo de log
  - Armazenar em `/var/lib/mini-docker/containers/<id>/output.log`
  - Exibir o conteúdo com `mini-docker logs <id>`

---

## Fase 6 — Variáveis de Ambiente e Configuração

- [ ] **#032 — Implementar flag `-e` para variáveis de ambiente**
  - `mini-docker run -e KEY=VALUE alpine env`
  - Passar variáveis via `Cmd.Env` no `os/exec`
  - Suportar múltiplas flags `-e`

- [ ] **#033 — Implementar flag `--hostname`**
  - `mini-docker run --hostname meu-container alpine`
  - Chamar `syscall.Sethostname()` dentro do container

- [ ] **#034 — Implementar flag `--read-only`**
  - Montar o rootfs como read-only
  - Montar `/tmp` como tmpfs para escrita temporária

---

## Fase 7 — Melhorias e Polimento

- [ ] **#035 — Implementar signal forwarding**
  - Capturar SIGINT/SIGTERM no processo pai
  - Repassar para o processo filho (PID 1 do container)
  - Implementar graceful shutdown

- [ ] **#036 — Implementar DNS no container**
  - Copiar ou gerar `/etc/resolv.conf` dentro do container
  - Apontar para o DNS do host ou para `8.8.8.8`

- [ ] **#037 — Implementar seccomp (opcional, avançado)**
  - Restringir syscalls perigosas (ex: `reboot`, `mount`, `kexec_load`)
  - Usar o pacote `libseccomp-golang`
  - Aplicar um perfil default similar ao do Docker

- [ ] **#038 — Implementar capabilities (opcional, avançado)**
  - Dropar Linux capabilities desnecessárias
  - Manter apenas: `CAP_NET_BIND_SERVICE`, `CAP_CHOWN`, etc.
  - Usar o pacote `syndtr/gocapability`

- [ ] **#039 — Escrever testes automatizados**
  - Testes unitários para parsing de configuração
  - Testes de integração que rodam containers de verdade (requer root)
  - Usar `testing` package do Go + test helpers

- [ ] **#040 — Documentar o projeto**
  - README.md com arquitetura, diagrama, e instruções de uso
  - Documentar cada flag e comando
  - Adicionar exemplos de uso e GIFs/screenshots

---

## Mapa de Dependências entre Issues

```
Fase 0: #001 → #002 → #003 → #004
                                 │
Fase 1: #005 → #006 → #007 → #008 → #009 → #011
                                              │
Fase 2: ──────── #012 → #013 → #014 → #015 → #016
                                              │
Fase 3: ──────────────── #017 → #018 → #019 → #020 → #021
                                                       │
Fase 4: ────────────────────── #022 → #023 → #024 → #025
                                                       │
Fase 5: ──── #026 → #027 → #028 → #029 → #030 → #031
                                                   │
Fase 6: ────────────────── #032 → #033 → #034      │
                                                   │
Fase 7: ──── #035 → #036 → #037 → #038 → #039 → #040
```

---

## Recursos Recomendados

### Leituras obrigatórias
- "Containers from Scratch" — Talk da Liz Rice (YouTube + código Go)
- "Linux Containers in 500 Lines of Code" — artigo
- Man pages: `namespaces(7)`, `cgroups(7)`, `pivot_root(2)`, `clone(2)`

### Repositórios de referência
- `lizrice/containers-from-scratch` — Implementação mínima em Go
- `p8952/bocker` — Docker implementado em ~100 linhas de bash
- `opencontainers/runc` — O runtime real do Docker (complexo, mas boa referência)

### Documentação
- Kernel docs: `Documentation/cgroup-v2.txt`
- OCI Runtime Spec: `github.com/opencontainers/runtime-spec`
- Go syscall package: `pkg.go.dev/syscall`

### Livros
- "Container Security" — Liz Rice (O'Reilly)
- "Linux System Programming" — Robert Love

---

## Checklist Final — Funcionalidades do Mini-Docker

Quando todas estas features estiverem funcionando, seu mini-docker está completo:

- [ ] `mini-docker run <image> <cmd>` — cria e roda um container
- [ ] `mini-docker run -e VAR=val` — variáveis de ambiente
- [ ] `mini-docker run -p 8080:80` — port mapping
- [ ] `mini-docker run --hostname x` — hostname customizado
- [ ] `mini-docker run --read-only` — filesystem read-only
- [ ] `mini-docker ps` — lista containers em execução
- [ ] `mini-docker exec <id> <cmd>` — executa comando em container existente
- [ ] `mini-docker stop <id>` — para um container
- [ ] `mini-docker rm <id>` — remove um container
- [ ] `mini-docker logs <id>` — mostra logs do container
- [ ] `mini-docker pull <image>` — baixa rootfs de uma imagem
- [ ] Container isolado com namespaces (UTS, PID, Mount, Net, IPC)
- [ ] Filesystem isolado com pivot_root + OverlayFS
- [ ] Recursos limitados com cgroups (memória, CPU, PIDs)
- [ ] Networking funcional com bridge + veth + NAT
