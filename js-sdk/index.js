class Config
{
    /**
     * @param {string} host - domain or IP address, e.g. 'localhost' 
     * @param {string} port - port number, e.g. 8372
     * @param {Boolean} verbose - should use versbose logging?
     */
    constructor(host, port, verbose) {
        this.host = '' + host;
        this.port = '' + port;
        this.verbose = Boolean(verbose)
    }

    /**
     * Returns URL address part for instance.
     * @return {string}
     */
    getAddress() {
        return `${this.host}:${this.port}`;
    }

    /**
     * Returns URL for hub.
     * @return {string}
     */
    getControlUrl() {
        return `ws://${this.host}:${this.port}/ctl`;
    }

    /**
     * Returns URL for instance.
     * @return {string}
     */
    getKeyValueApiUrl(suffix) {
        return `ws://${this.host}:${this.port}/ws`;
    }

    /**
     * Returns true if logging should be verbose.
     * @return {Boolean}
     */
    isVerbose()
    {
        return this.verbose;
    }
}

class BaseApiClient
{
    /**
     * @param {string} url
     * @param {boolean} verbose
     */
    constructor(url, verbose) {
        this.url = url;
        this.socket = new WebSocket(url);
        this.verbose = verbose;
        this.requestId = 0;
        this.requestMapping = {};

        this.socket.onopen = this._onopen.bind(this);
        this.socket.onclose = this._onclose.bind(this);
        this.socket.onmessage = this._onmessage.bind(this);
        this.socket.onerror = this._onerror.bind(this);
        this.pendingSend = [];
        this.isOpen = false;
    }

    /**
     * Sends request if socket is open, otherwise adds to pending requests.
     * @param {string} action 
     * @param {string} option1 
     * @param {string} option2 
     */
    sendRequest(action, option1 = '', option2 = '') {
        return new Promise((resolve, reject) => {
            let requestId = ++this.requestId;
            this.requestMapping[requestId] = {
                'resolve': resolve,
                'reject': reject
            };
            let send = this._sendRequestBySocket.bind(
                this, requestId, action, option1, option2);
            if (this.isOpen)
            {
                send();
            }
            else
            {
                this.pendingSend.push(send);
            }
        });
    }

    _onopen() {
        this.isOpen = true;
        this._sendAllPending();
        this._log('connection established with ', this.url);
    }

    _sendAllPending()
    {
        for (let index in this.pendingSend)
        {
            this.pendingSend[index]();
        }
        this.pendingSend = [];
    }

    _onclose(event) {
        this.isOpen = false;
        const status = event.wasClean ? 'closed' : 'aborted';
        this._log('connection ', status, 'code: ', event.code, ', reason: ', event.reason);
    }

    _sendRequestBySocket(requestId, action, option1 = '', option2 = '') {
        let myObj = {
            'Payload': JSON.stringify({
                'Action': '' + action,
                'Option1': '' + option1,
                'Option2': '' + option2
            }),
            'RequestId': requestId
        };
        this.socket.send(JSON.stringify(myObj));
    }

    _parseResponse(data) {
        const response = JSON.parse(data);
        const requestId = response['RequestId'];
        if (requestId in this.requestMapping)
        {
            const payload = JSON.parse(response['Payload']);
            const handlers = this.requestMapping[requestId];
            if (Boolean(payload['Success']))
            {
                handlers.resolve(payload['Result']);
            }
            else
            {
                handlers.reject('' + payload['Error']);            
            }
            delete handlers[requestId];
        }
    }

    _onmessage(event) {
        this._log('got data: ', event.data);
        this._parseResponse(event.data);
    }

    _onerror(error) {
        this._log('got error: ', error);
    }

    _log(...args) {
        if (this.verbose)
        {
            console.log(...args);
        }
    }
}

class HubApiClient extends BaseApiClient
{
    /**
     * Creates admin client to control Hub.
     * @param {!Config} config
     */
    constructor(config) {
        super(config.getControlUrl(), config.isVerbose());
        this.config = config;
    }
    
    /**
     * Lists all running instances
     * Returns array of strings
     */
    list() {
        return this.sendRequest('LIST').then((data) => {
            const items = JSON.parse(data);
            if (items instanceof Array) {
                return items;
            }
            throw new Error('internal error: LIST command returned value of type ' + (typeof items));
        });
    }

    /**
     * Checks if instance on given port exists, run a new one if doesn't.
     * @param {String} port 
     * @returns {Promise<InstanceApiClient>}
     */
    get(port) {
        let instanceConfig = new Config(this.config.host, port, this.config.isVerbose());
        return this.list().then((addressList) => {
            if (addressList.indexOf(instanceConfig.getAddress()) >= 0) {
                return new InstanceApiClient(instanceConfig);
            }
            return this.run(port);
        });
    }

    /**
     * Runs new key-value storage instance on URL with given suffix.
     * @param {String} port 
     * @returns {Promise<InstanceApiClient>}
     */
    run(port) {
        let instanceConfig = new Config(this.config.host, port, this.config.isVerbose());
        return this.sendRequest('SET', instanceConfig.getAddress()).then(() => {
            return new InstanceApiClient(instanceConfig);
        });
    }
    
    /**
     * Stops key-value storage instance.
     * @param {String} suffix 
     */
    stop(suffix) {
        const url = this.config.getInstanceUrl(suffix);
        return this.sendRequest('REMOVE', url).then(() => {});
    }
}

class InstanceApiClient extends BaseApiClient
{
    /**
     * Creates instance client which provides key-value storage.
     * @param {!Config} config
     */
    constructor(config) {
        super(config.getKeyValueApiUrl(), config.isVerbose());
    }

    /**
     * Lists all keys and values in storage
     * @returns {Object} dictionary which maps keys to values.
     */
    list() {
        return this.sendRequest('LIST').then((data) => {
            const items = JSON.parse(data);
            if (items instanceof Object) {
                return items;
            }
            throw new Error('internal error: LIST returned value of type ' + (typeof items));
        });
    }

    /**
     * Puts key/value pair to storage
     * @param {string} key 
     * @param {string} value 
     */
    set(key, value) {
        return this.sendRequest('SET', key, value).then(() => {});
    }

    /**
     * Reads stored value for given.
     * @param {string} key 
     */
    get(key) {
        return this.sendRequest('GET', key).then((value) => {
            this._log('GET response: ', value);
            if (value instanceof String) {
                return value;
            }
            throw new Error('internal error: GET returned value of type ' + (typeof items));
        });
    }

    /**
     * Removes value for given key.
     * @param {string} key 
     */
    remove(key) {
        return this.sendRequest('REMOVE', key).then(() => {});
    }
}
