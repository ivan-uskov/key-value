
class Editor {
    /**
     * Creates editor, shows editor UI on HTML page.
     * @param {InstanceApiClient} client
     */
    constructor(client) {
        this.client = client;
        this.connectButtons();
        this.client.addConnectionUpdatedHandler(this.updateTable.bind(this));
        this.updateTable();
    }

    connectButtons() {
        let addButton = document.getElementById('editor-add-form-add');
        let removeButton = document.getElementById('editor-add-form-remove');
        let keyInput = document.getElementById('editor-add-form-key');
        let valueInput = document.getElementById('editor-add-form-value');
        addButton.addEventListener('click', () => {
            let key = keyInput.value;
            let value = valueInput.value;
            if (key.length !== 0 && value.length !== 0)
            {
                this.addKeyValue(key, value);
            }
        });
        removeButton.addEventListener('click', () => {
            let key = keyInput.value;
            if (key.length !== 0)
            {
                this.removeKey(key);
            }
        });
    }

    async addKeyValue(key, value) {
        await this.client.set(key, value);
        this.updateTable();
    }

    async removeKey(key) {
        await this.client.remove(key);
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
        for (let key of Object.keys(data)) {
            keyColumn.textContent = '' + key;
            valueColumn.textContent = '' + data[key];
            let row = document.importNode(rowTemplate.content, true);
            tbody.appendChild(row);
        }

        // Force MDL to update - table has new elements.
        window.componentHandler.upgradeAllRegistered();
    }
}

function hideLoadSpinner()
{
    let spinner = document.querySelector('#editor-load-spinner');
    spinner.style.visibility = 'hidden';
    spinner.style.display = 'none';
}

function getInstanceApiClient() {
    let config = new Config('localhost', '8372', true);
    let hub = new HubApiClient(config);
    return hub.get('8375');
}

function showContent() {
    hideLoadSpinner();

    let grid = document.querySelector('#editor-content-grid');
    grid.style.visibility = 'visible';
    grid.style.display = 'block';
}

async function runEditor() {
    try
    {
        let client = await getInstanceApiClient();
        showContent();
        window.editor = new Editor(client);
    }
    catch (err)
    {
        alert('' + err);
        hideLoadSpinner();
    }
}

runEditor();
