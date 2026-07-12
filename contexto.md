# Planejamento Técnico: Migração de React para Vue 3 (Vite)

## 1. Visão Geral e Objetivo
O objetivo é criar uma interface frontend oficial para o Identity Provider `nayz-auth` (Admin Console e Telas de Login) utilizando **Vue 3 + Vite**. O novo sistema irá migrar e evoluir o modelo conceitual (`draft`) atualmente escrito em React, sendo armazenado na nova pasta raiz chamada `/frontend`.

A escolha do Vue 3 baseia-se na sua leveza incomparável, reatividade profunda (Composition API) e simplicidade de código em relação ao React (sem necessidade de Hooks complexos, useCallback ou gerenciamento excessivo de renderizações).

## 2. Stack Tecnológica
*   **Framework Base:** Vue 3 (Composition API via `<script setup>` com TypeScript)
*   **Bundler:** Vite (Altíssima performance de build e HMR)
*   **Roteamento:** `vue-router` (Padrão oficial da comunidade)
*   **Estilização e UI:** Tailwind CSS. Para componentes ricos (ex: Tabs, InputOTP do draft React), utilizaremos os equivalentes portados para o ecossistema Vue (como `shadcn-vue` / `radix-vue`) ou implementações customizadas super limpas com Headless UI.
*   **Ícones:** `lucide-vue-next`
*   **Notificações (Toasts):** `vue-sonner` (port exato do Sonner utilizado no React).
*   **Comunicação com API:** Fetch nativo estruturado ou Axios (reescrevendo o atual `lib/api.ts`).

## 3. Correção Crítica de UX e Arquitetura (A Tela de Login)
No *draft* atual (React), a tela de Login (`auth.login.tsx`) exige que o usuário preencha manualmente o ID da aplicação através de um input visual (`AppIdField`).
**O Ajuste:** Como este Frontend será o painel administrativo **oficial e exclusivo** da própria plataforma `nayz-auth` (para acesso do usuário Root), o administrador não deveria ter que informar qual é o UUID do sistema.
*   O `app_id` oficial mestre será definido como uma Variável de Ambiente (`VITE_NAYZ_AUTH_APP_ID`) no arquivo `.env` do frontend.
*   Toda requisição de autenticação (Login Clássico e Início do Fluxo Passwordless) injetará esse UUID "por debaixo dos panos", abstraindo completamente essa complexidade.
*   O componente `AppIdField` será removido da tela de Login, deixando a interface idêntica às ferramentas Cloud tradicionais (apenas E-mail e Senha/Código).

## 4. Etapas de Execução Propostas

### Etapa 1: Setup do Ambiente e Estrutura
1. Rodar `npx create-vite@latest frontend --template vue-ts` na raiz do repositório.
2. Instalar dependências base de roteamento, UI e HTTP.
3. Configurar caminhos no `vite.config.ts` (alias `@/`) e integrar o Tailwind CSS.

### Etapa 2: Lógica Base de Autenticação (Serviços e Router)
1. Traduzir o arquivo `lib/api.ts` do React para um service TypeScript no Vue.
2. Implementar a injeção do `VITE_NAYZ_AUTH_APP_ID`.
3. Criar o Roteador (`router.ts`) definindo *Navigation Guards*:
   *   Qualquer acesso à rota `/admin` ou seus filhos só será permitido se existir um `token` JWT armazenado no `localStorage`. Caso contrário, ocorrerá o redirecionamento automático para `/auth/login`.

### Etapa 3: Componentização Visual e Telas
1. **Layout e Theme:** Trazer o CSS Global e a estratégia de Tema Escuro/Claro (Dark Mode).
2. **Shell de Autenticação:** Construir o componente `AuthShell.vue` que abraça os formulários de entrada.
3. **Página de Login (`Login.vue`):** 
   *   Migrar o controle de formulários para a reatividade do Vue (`ref`, `reactive`).
   *   Implementar a funcionalidade de Tabs (Alternância entre Login com Senha vs Login com Código OTP via E-mail).
4. **Página de Registro (`Register.vue`):** Implementar formulário básico de registro (caso o root decida expor a criação de novas contas abertamente).
5. **Dashboard Administrativo (`Admin.vue`):**
   *   Criar a estrutura base do painel (Sidebar / Header).
   *   Criar telas/tabelas para listar, criar, atualizar e deletar **Aplicações** (chamando os endpoints REST criados anteriormente no Go).
   *   Criar tela para gestão de **Roles** e vinculação aos **Usuários**.

---
**Ação Necessária:** O planejamento acima cobre detalhadamente os ajustes solicitados e o mapeamento para migrar de React para Vue 3. Se estiver de acordo, retorne a confirmação para darmos início técnico gerando a pasta `frontend` e iniciando os códigos.