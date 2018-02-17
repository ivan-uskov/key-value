class Config {
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
    getKeyValueApiUrl() {
        return `ws://${this.host}:${this.port}/ws`;
    }

    /**
     * Returns true if logging should be verbose.
     * @return {Boolean}
     */
    isVerbose() {
        return this.verbose;
    }
}

const RECONNECTION_TIMEOUT = 1000;

class BaseApiClient {
    /**
     * @param {string} url
     * @param {boolean} verbose
     */
    constructor(url, verbose) {
        this.url = url;
        this.verbose = verbose;
        this.connectionTimes = 0;
        this.connectionUpdatedHandler = null;
        this._createConnection();
    }

    /**
     * @param h () => void
     */
    addConnectionUpdatedHandler(h) {
        this.connectionUpdatedHandler = h;
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
            if (this.isOpen) {
                send();
            }
            else {
                this.pendingSend.push(send);
            }
        });
    }

    /**
     * @protected
     */
    _createConnection() {
        this.requestId = 0;
        this.requestMapping = {};
        this.socket = new WebSocket(this.url);

        this.socket.onopen = this._onopen.bind(this);
        this.socket.onclose = this._onclose.bind(this);
        this.socket.onmessage = this._onmessage.bind(this);
        this.socket.onerror = this._onerror.bind(this);
        this.pendingSend = [];
        this.isOpen = false;
        ++this.connectionTimes;
    }

    _onopen() {
        this.isOpen = true;
        this._sendAllPending();
        this._log('connection established with ', this.url);
    }

    _sendAllPending() {
        for (let value of this.pendingSend) {
            value();
        }
        this.pendingSend = [];
    }

    _sendRequestBySocket(requestId, action, option1 = '', option2 = '') {
        let myObj = {
            'payload': JSON.stringify({
                'action': '' + action,
                'option_1': '' + option1,
                'option_2': '' + option2
            }),
            'request_id': requestId
        };
        this.socket.send(JSON.stringify(myObj));
    }

    _parseResponse(data) {
        const response = JSON.parse(data);
        const requestId = response['request_id'];
        if (requestId in this.requestMapping) {
            const payload = JSON.parse(response['payload']);
            const handlers = this.requestMapping[requestId];
            if (Boolean(payload['success'])) {
                handlers.resolve(payload['result']);
            }
            else {
                handlers.reject(new Error('' + payload['error']));
            }
            delete handlers[requestId];
        }
    }

    _rejectAll(message) {
        for (let handlers of Object.values(this.requestMapping)) {
            handlers.reject(new Error(message));
        }
        this.requestMapping = {};
    }

    _onmessage(event) {
        this._log('got data: ', event.data);
        this._parseResponse(event.data);
    }

    _onerror(error) {
        // With WebSockets, onerror is always followed by termination of connection.
        this._log('got error: ', error);
    }

    _onclose(event) {
        this._rejectAll('connection closed with code ' + event.code);

        this.isOpen = false;
        const status = event.wasClean ? 'closed' : 'aborted';
        this._log('connection ' + status + ', event: ', event);
        setTimeout(() => {
            this._createConnection();
            this.connectionUpdatedHandler && this.connectionUpdatedHandler();
        }, RECONNECTION_TIMEOUT);
    }

    _log(...args) {
        if (this.verbose) {
            console.log(...args);
        }
    }
}

class HubApiClient extends BaseApiClient {
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
        return this._run(port);
    }

    /**
     * Stops key-value storage instance.
     * @param {String} suffix
     */
    stop(suffix) {
        const url = this.config.getInstanceUrl(suffix);
        return this.sendRequest('REMOVE', url).then(() => {
        });
    }

    /**
     * Runs new key-value storage instance on URL with given suffix.
     * @param {String} port
     * @returns {Promise<InstanceApiClient>}
     */
    _run(port) {
        let instanceConfig = new Config(this.config.host, port, this.config.isVerbose());
        return this.sendRequest('RUN', instanceConfig.getAddress()).then(() => {
            return new InstanceApiClient(instanceConfig);
        });
    }
}

class InstanceApiClient extends BaseApiClient {
    /**
     * Creates instance client which provides key-value storage.
     * @param {!Config} config
     */
    constructor(config) {
        super(config.getKeyValueApiUrl(), config.isVerbose());

        this.cofig = config;
    }

    getConfig() {
        return this.cofig;
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
        return this.sendRequest('SET', key, value).then(() => {
        });
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
        return this.sendRequest('REMOVE', key).then(() => {
        });
    }
}
