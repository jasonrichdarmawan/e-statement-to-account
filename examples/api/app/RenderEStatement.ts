const estatementElement = document.querySelector("div[id='EStatement']")
if (estatementElement === null) {
    throw Error("can't find element with selector: div[id='EStatement']")
}
const transactionsElement = <HTMLDivElement | undefined>estatementElement.children[0];
const groupedByAccountElement = <HTMLDivElement | undefined>estatementElement.children[1];
const accountSummaryElement = <HTMLDivElement | undefined>estatementElement.children[2];

interface Transaction {
    Date: string;
    Description1: string | null;
    Description2: string | null;
    Branch: string | null;
    Mutation: number;
    Entry: string | null;
    Balance: number;
}

interface Accounts {
    AccountNames: string[]
    Transactions: Transaction[][]
    Balances: number[]
}

interface ParserResponse {
    Transactions: Transaction[] | null;
    Accounts: Accounts | null;
}

function getTransactionsTBodiesElement(): HTMLTableSectionElement {
    if (transactionsElement === undefined) {
        throw Error("element can't be null")
    }

    const panelElement = transactionsElement.querySelector("div[id='panel']")
    if (panelElement == null) {
        throw Error("can't find element with selector: div[id='panel']")
    }

    const tableElement = <HTMLTableElement>panelElement.children[0]

    return tableElement.tBodies[0]
}

function RenderTransactions(response: ParserResponse) {
    if (response.Transactions == null) {
        throw Error("The Transactions key from the response body should not be empty")
    }

    if (transactionsElement === undefined) {
        throw Error("element can't be null")
    }
    transactionsElement.hidden = false

    let tBodies = getTransactionsTBodiesElement()

    tBodies.innerHTML = ""

    response.Transactions.forEach((transaction) => {
        const newRowElement = tBodies.insertRow()
        for (let i = 0; i < 7; i++) {
            const newCellElement = newRowElement.insertCell()
            switch (i) {
                case 0:
                    newCellElement.innerHTML = atob(transaction.Date)
                    break;
                case 1:
                    if (transaction.Description1 == null) {
                        break;
                    }
                    newCellElement.innerHTML = atob(transaction.Description1)
                    break;
                case 2:
                    if (transaction.Description2 == null) {
                        break;
                    }
                    newCellElement.innerHTML = atob(transaction.Description2)
                    break;
                case 3:
                    if (transaction.Branch == null) {
                        break;
                    }
                    newCellElement.innerHTML = atob(transaction.Branch)
                    break;
                case 4:
                    if (transaction.Mutation == null || transaction.Mutation == 0) {
                        break;
                    }
                    newCellElement.innerHTML = Number(transaction.Mutation).toLocaleString()
                    break;
                case 5:
                    if (transaction.Entry == null) {
                        break;
                    }
                    newCellElement.innerHTML = atob(transaction.Entry)
                    break;
                case 6:
                    if (transaction.Balance == null || transaction.Balance == 0) {
                        break;
                    }
                    newCellElement.innerHTML = Number(transaction.Balance).toLocaleString()
                    break;
            }
        }
    })
}

function getGroupedByAccountPanelElement(): HTMLDivElement {
    if (groupedByAccountElement === undefined) {
        throw Error("element can't be null")
    }

    const panelElement = <HTMLDivElement>groupedByAccountElement.querySelector("div[id='panel']")
    if (panelElement == null) {
        throw Error("can't find element with selector: div[id='panel']")
    }

    return panelElement
}

function RenderGroupedByAccount(response: ParserResponse) {
    if (response.Accounts == null) {
        throw Error("The Transactions key from the response body should not be empty")
    }

    if (groupedByAccountElement === undefined) {
        throw Error("element can't be null")
    }
    groupedByAccountElement.hidden = false

    let panelElement = getGroupedByAccountPanelElement()

    panelElement.innerHTML = ""

    for (let x = 0; x < response.Accounts.AccountNames.length; x++) {
        const tableElement = document.createElement("table")

        const tHeadElement = tableElement.createTHead()
        const tHeadRow1Element = tHeadElement.insertRow()
        tHeadRow1Element.insertCell().outerHTML = `<th colspan="3">${atob(response.Accounts.AccountNames[x])}</th>`
        const tHeadRow2Element = tHeadElement.insertRow()
        tHeadRow2Element.insertCell().outerHTML = "<th>TANGGAL</th>"
        tHeadRow2Element.insertCell().outerHTML = "<th>KETERANGAN</th>"
        tHeadRow2Element.insertCell().outerHTML = "<th>MUTASI</th>"

        const tBodyElement = tableElement.createTBody()
        for (let y = 0; y < response.Accounts.Transactions[x].length; y++) {
            const transaction = response.Accounts.Transactions[x][y]

            const tBodyRow1Element = tBodyElement.insertRow()
            tBodyRow1Element.insertCell().innerHTML = atob(transaction.Date)
            if (transaction.Description2 !== null) {
                tBodyRow1Element.insertCell().innerHTML = atob(transaction.Description2)
            } else if (transaction.Description1 !== null) {
                tBodyRow1Element.insertCell().innerHTML = atob(transaction.Description1)
            }
            const mutasi = transaction.Mutation === 0 ? "" : Number(transaction.Mutation).toLocaleString()
            const entry = transaction.Entry === null ? "" : ` ${atob(transaction.Entry)}`
            tBodyRow1Element.insertCell().innerHTML = mutasi + entry
        }

        panelElement.appendChild(tableElement)
    }
}

function getAccountSummaryTBodiesElement(): HTMLTableSectionElement {
    if (accountSummaryElement === undefined) {
        throw Error("element can't be null")
    }

    const panelElement = accountSummaryElement.querySelector("div[id='panel']")
    if (panelElement == null) {
        throw Error("can't find element with selector: div[id='panel']")
    }

    const tableElement = <HTMLTableElement>panelElement.children[0]

    return tableElement.tBodies[0]
}

function RenderAccountSummary(response: ParserResponse) {
    if (response.Accounts == null) {
        throw Error("The Transactions key from the response body should not be empty")
    }

    if (accountSummaryElement === undefined) {
        throw Error("element can't be null")
    }
    accountSummaryElement.hidden = false

    const tBodiesElement = getAccountSummaryTBodiesElement()

    tBodiesElement.innerHTML = ""

    for (let i = 0; i < response.Accounts.AccountNames.length; i++) {
        const newRow = tBodiesElement.insertRow()
        newRow.insertCell().innerHTML = atob(response.Accounts.AccountNames[i])
        newRow.insertCell().innerHTML = Number(response.Accounts.Balances[i]).toLocaleString()
    }
}

function CleanAllRender() {
    if (transactionsElement === undefined || groupedByAccountElement === undefined || accountSummaryElement === undefined) {
        throw Error("element can't be null")
    }
    transactionsElement.hidden = true
    getTransactionsTBodiesElement().innerHTML = ""
    groupedByAccountElement.hidden = true
    getGroupedByAccountPanelElement().innerHTML = ""
    accountSummaryElement.hidden = true
    getAccountSummaryTBodiesElement().innerHTML =""
}