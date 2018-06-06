
class Editor {
    /**
     * Creates editor, shows editor UI on HTML page.
     * @param {InstanceApiClient} client
     */
    constructor(client) {
        this.client = client;
        this.initializeEditor();
        this.connectButtons();
        this.client.addConnectionUpdatedHandler(this.updateTable.bind(this));
        this.updateTable();

        setInterval(this.updateTable.bind(this), 1000);
    }

    initializeEditor() {
        let rowTemplate = document.querySelector('#instance_values');
        let row = document.importNode(rowTemplate.content, true);
        let container = document.querySelector('#editor-content-grid');
        container.prepend(row);
        this.head = container.firstElementChild;
        this.head.querySelector('.main_title').innerHTML = this.client.getConfig().getKeyValueApiUrl();
    }

    connectButtons() {
        let addButton = this.head.querySelector('.editor-add-form-add');
        let removeButton = this.head.querySelector('.editor-add-form-remove');
        let keyInput = this.head.querySelector('.editor-add-form-key');
        let valueInput = this.head.querySelector('.editor-add-form-value');
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
        let table = this.head.querySelector('.editor-key-value-table');
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

function showContent() {
    hideLoadSpinner();

    let grid = document.querySelector('#editor-content-grid');
    grid.style.visibility = 'visible';
    grid.style.display = 'block';
}

async function runEditor() {
    try
    {
        let config = new Config('localhost', '8372', true);
        let hub = new HubApiClient(config);

        const count = 10;
        for (let i = 0; i < count; ++i)
        {
            const num = i;
            const client = await hub.get('' + (8376 + num));
            window['editor' + num] = new Editor(client);
        }

        showContent();
    }
    catch (err)
    {
        alert('' + err);
        hideLoadSpinner();
    }
}

runEditor();
