# Aplicação de Mensagens em Tempo Real

Esta aplicação é um sistema de mensagens em tempo real, projetado para gerenciar comunicação entre múltiplos usuários de maneira eficiente e escalável, utilizando WebSockets.

## Recursos Principais

- **Mensagens em Tempo Real**  
  Os usuários podem enviar e receber mensagens instantaneamente, com cada mensagem contendo remetente, destinatário e conteúdo.

- **Comunicação Persistente com WebSockets**  
  A aplicação mantém conexões abertas com WebSockets, permitindo um fluxo contínuo de mensagens sem a necessidade de múltiplas requisições HTTP.

- **Gerenciamento de Conexões**  
  Um `ConnectionManager` gerencia as conexões WebSocket ativas, associando-as ao ID de cada usuário, facilitando a entrega de mensagens ao destinatário correto.

- **Pool de Workers para Processamento**  
  O sistema utiliza um pool de workers para processar mensagens simultaneamente, distribuindo a carga de trabalho de envio e aumentando o desempenho geral.

- **Processamento em Lote**  
  Mensagens são agrupadas em lotes antes de serem enviadas, otimizando o uso de recursos ao reduzir operações de I/O.

- **Controle de Inatividade**  
  Conexões inativas são monitoradas e encerradas automaticamente após um período sem resposta, garantindo a liberação de recursos.

- **Escalabilidade e Resiliência**  
  Utilizando canais de mensagens (`jobQueue`), a aplicação suporta alta demanda sem sobrecarregar o servidor, garantindo que as mensagens sejam processadas de forma eficiente.

## Casos de Uso

Este sistema pode ser utilizado e adaptado para:

- Aplicativos de mensagens instantâneas
- Notificações em tempo real
- Sistemas de comunicação interna em equipes

Com essas funcionalidades, a aplicação oferece uma base para sistemas de chat em tempo real.
