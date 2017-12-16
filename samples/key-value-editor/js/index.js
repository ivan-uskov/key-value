
class Editor {
    /**
     * Creates editor, shows editor UI on HTML page.
     * @param {InstanceApiCliet} client 
     */
    constructor(client) {
        console.log('started Editor ctor');
        this.client = client;
        this.showContent();
        this.connectButtons();
        this.updateTable();
        console.log('finished Editor ctor');
    }

    showContent() {
        let spinner = document.querySelector('#editor-load-spinner');
        let grid = document.querySelector('#editor-content-grid');
        spinner.style.visibility = 'hidden';
        grid.style.visibility = 'visible';
    }
    
    connectButtons() {
        let addButton = document.getElementById('editor-add-form-submit');
        let keyInput = document.getElementById('editor-add-form-key');
        let valueInput = document.getElementById('editor-add-form-value');
        addButton.addEventListener('click', (event) => {
            let key = keyInput.value;
            let value = valueInput.value;
            if (key.length != 0 && value.length != 0)
            {
                this.submitKeyValue(key, value);
            }
        });
    }

    async submitKeyValue(key, value) {
        await this.client.set(key, value);
        this.updateTable();
    }
    
    async updateTable() {
        // Query table nodes
        let table = document.querySelector('#editor-key-value-table');
        let tbody = table.querySelector("tbody");

        // Query template nodes
        let rowTemplate = document.querySelector('#editor-key-value-row');
        let columns = rowTemplate.content.querySelectorAll('td');
        let keyColumn = columns[0];
        let valueColumn = columns[1];

        let data = await this.client.list();

        // Cleanup and create table again.
        while (tbody.firstChild) {
            tbody.removeChild(tbody.firstChild);
        }
        for (let key in data) {
            keyColumn.textContent = '' + key;
            valueColumn.textContent = '' + data[key];
            let row = document.importNode(rowTemplate.content, true);
            tbody.appendChild(row);
        }
    }
}

function getInstanceApiClient() {
    let config = new Config('localhost', '8372', true);
    let hub = new HubApiClient(config);
    return hub.get('8375');
}

async function runEditor() {
    let client = await getInstanceApiClient();
    window.editor = new Editor(client);
}

runEditor()
