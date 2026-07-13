DIARIO DE BORDO 
- anotações diárias de progresso do projeto

DIA 01(LEARN - 01/05/2026):
Feito hoje: pesquisei como funciona o SPI, o DICT, mensagens ISO 20022; aprendi conceitos EndToEndId e protocolos que vêm por meio de xml(irei usar json). 
Modelei as entidades no drawSQL e a relação entre as tabelas

Amanhã: Para o Dia 2 máquina de estados das transferências

DIA 02(BUILD - 02/05/2026):
Feito hoje: criada relação das pastas; criado arquivo docker-compose com postgres 16 e arquivo .env com as credencias(adicionado ao gitignore).
Essas credencias são criadas automaticamente pelo Docker quando rodar ```docker-compose up -d```- o Postgres inicializa com elas.
Criada conexão com o postgres.

Amanhã: testar conexão

dia 03(BUILD - 03/05/2026):
Feito hoje: testada a conexão com o postgres utilizando a biblioteca pgx e teste conn.Ping()
    - dificuldades encontradas: a porta 5432 estava sendo ocupada por uma versão anterior do postgres
    - migrei para a porta 5433
    - para o docker sempre o host será postgres e a porta 5432 independente
Amanhã: migrations

dia 05(Build - 05/05/2026):
feito hoje:
    -instalei o golang migrate(cli)
        - go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    - criei arquivos de migration
        migrate create -ext sql -dir migrations -seq create_users
        migrate create -ext sql -dir migrations -seq create_participants
        migrate create -ext sql -dir migrations -seq create_accounts
        migrate create -ext sql -dir migrations -seq create_transfers
        migrate create -ext sql -dir migrations -seq create_ledger_entries
        migrate create -ext sql -dir migrations -seq create_pix_keys
    - o que é o golang-migrate?
        - arquivos que controlam a evolução do banco de dados ao longo do tempo
        - golang-migrate executa os sqls em ordem
        - O golang-migrate guarda em uma tabela interna (schema_migrations) qual foi a última migration executada, então ele sabe de onde continuar.
    - por que é útil?
        - sem migration: toda vez que o banco resetar(docker-compose down -v) seria preciso recriar as tabelas na mão.
        - com migration: um comando recria tudo:
            - migrate ... up
        - cada mudança vira uma nova migration numerada, e o histórico fica no git.
    - por que .up e .down?
        .up.sql aplica a mudança(create table)
        .down.sql desfaz a mudança(drop table)
    - .down volta commits
        migrate ... down 2   # desfaz as 2 últimas migrations
        migrate ... up       # reaaplica
    - roda o comando pra criar os migrations:
        migrate -path migrations -database "postgresql://pay_user:pay_pass@localhost:5433/project_pay?sslmode=disable" up

    - onde ver as tabelas?
        project_pay
        └── Schemas
                └── public
                    └── Table

    para amanhã: POST /accounts GET /accounts/{id}
    go.mod — corrigir o nome do módulo (module main.go → module project_pay)
    cmd/internal/database/database.go — mover a lógica de conexão do client/main.go para cá
    cmd/internal/models/users.go — struct Account
    cmd/internal/repositories/accounts/accounts.go — queries SQL
    cmd/internal/handlers/accounts.go — handler HTTP
    cmd/api/main.go — roteador e servidor

dia 07(Build - 07/05/2026)
    feito hoje:
        - corrigido nome do módulo
        -  cmd/internal/models/users.go


dia 08(Build - 15/05/2026)
    feito hoje:
        - corrigido nome do módulo
        - revisado as estruturas e métodos criados para users
        - criado as estrturas e métodos para accounts
    erros importantes:
        - só sabia montar as consultas em banco com a lib sql/db
        - aprendi a montar com pgx 
            - preferia sql/db é mais organizado(posso acabar migrando depois)
        - em postgres o getdate() do sqlserver é chamado por now()
    ficou para amanhã:
        - criar para participants
        - criar para ledger_entries
        - pix_keys 
        - handlers padrões
        - corrigir o diagrama com as relações corretas
dia 09(Build - 16/05/2026)
    feito hoje:
        - métodos e structs de participants, pix_keys, ledger_entries e uma parte de transfers
        -ajuste nas estruturas participants

    aprendido hoje:
        - ISPB = Identificador do Sistema de Pagamentos Brasileiro.
        - É um código de 8 dígitos que o Banco Central atribui para cada instituição financeira participante do sistema de pagamentos. É o "CPF do banco".

dia 10(Build - 17/05/2026)
    feito hoje:
        - finalizado a camada completa de transfers: model, repository com transação atômica, handler e rotas no main.go
        - implementada função generateEndToEndID no repositório
            formato: E + ISPB(8 dígitos) + YYYYMMDD + HHMM + sufixo aleatório(11 chars)
            exemplo: E60074319020260517142300A3F9C12D8
        - implementado ErrInsufficientFunds como erro exportado
            permite que o handler distinga saldo insuficiente(422) de erro interno(500)
        - fluxo da transação atômica:
            1. SELECT balance + ispb do remetente com FOR UPDATE (trava a linha)
            2. verifica saldo >= amount
            3. gera o end_to_end_id
            4. debita remetente (UPDATE ... RETURNING balance)
            5. verifica se conta recebedor existe
            6. credita recebedor (UPDATE ... RETURNING balance)
            7. INSERT na transfers
            8. INSERT ledger_entry DEBIT (remetente)
            9. INSERT ledger_entry CREDIT (recebedor)
            10. COMMIT
        - registradas rotas POST /transfers e GET /transfers/{id} no main.go

    lições aprendidas:
        - transfers é o único recurso que precisa de transação SQL pois são múltiplas operações que precisam ser atômicas
        - FOR UPDATE trava a linha do remetente durante a transação para evitar débitos simultâneos (race condition)
        - erros de negócio devem ser erros exportados (var ErrInsufficientFunds = errors.New(...)) para o handler conseguir tratá-los com errors.Is() e retornar o status HTTP correto
        - tx.Exec para operações sem retorno (ledger entries); tx.QueryRow quando precisa do RETURNING
        - defer tx.Rollback() é seguro chamar mesmo depois do Commit — pgx ignora se a transação já foi encerrada

    erros aprendidos:
        - CompletedAt era time.Time mas deve ser *time.Time pois o campo é nullable no banco; quando NULL o scan falha com time.Time
        - removido EndToEndID e Status do CreateTransferRequest: não devem vir do cliente, são gerados/controlados pelo sistema
        - sqlCreditAccount chamado com 1 argumento mas o SQL precisava de 2 ($1=amount, $2=id) — passou apenas o id
        - GetByID retornava transfer, nil ao invés de transfer, err após o scan
        - faltava verificação if err != nil após scan do recebedor — um erro de banco passaria silencioso

### RESUMO DO QUE FOI IMPLEMENTADO ATÉ AQUI
    No primeiro dia aprendi regras de negócio como SPI, o DICT, mensagens ISO 20022
    Modelei as entidades no drawSQL e a relação entre as tabelas
    Fiz a máquina de estado do programa
    criei as pastas e arquivos do repositório Project_Pay
    criei as tabelas com golang-migrate que são criadas toda vez via docker-compose
    criei uma conexão com o postgres utilizando a biblioteca pgx
    e todos os endpoints iniciais do projeto com as regras de validação para transferências atômicas

dia 10(REFACTOR/DOC - 21/05/2026)
    o que foi feito? 
    - Revisão do que foi criado até aqui
    - teste que primeira transferência usando o método POST /transfers ocorreu com sucesso

    lições devops: 
        pgAdmin - acessar
            URL:   http://localhost:5050
            Email: admin@admin.com
            Senha: admin
        
    Configuração do servidor no pgAdmin (fazer uma vez após down -v):
    Aba General  → Name: project_pay
    Aba Connection:
    Host:     host.docker.internal
    Port:     5432
    Database: project_pay
    Username: pay_user
    Password: pay_pass

    Banco — comandos úteis via Docker

    # acessar o banco direto pelo terminal
        docker exec -it project_pay_db psql -U pay_user -d project_pay

    # dentro do psql — comandos úteis
        \dt                  # listar tabelas
        \d nome_da_tabela   # ver estrutura da tabela
        \q                  # sair

    ORDEM PARA INICIAR DO ZERO
        1. PowerShell como admin → parar serviços PostgreSQL locais
        2. docker-compose up -d
        3. migrate ... up
        4. go run ./cmd/api/.
        5. Postman → testar endpoints


dia 10(BUILD - 22/05/2026)
    o que foi feito? 
        - criado autenticação JWT Auth
        - por meio de tokens bearer

dia 11(build )







    




