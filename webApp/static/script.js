initSomeGifs()
const expressions = document.querySelector('.expression-list');
const input = document.getElementById('user-input');

const nextButton = document.getElementById('send-button');
nextButton.onclick = e => handleUserInput(e);
input.onkeyup = e => e.key == "Enter" && handleUserInput(e);

function initSomeGifs() {
    const welcomeStaticPath = './assets/img/welcome_static.gif';
    const welcomePath = './assets/img/welcome.gif';
    const welcome = document.getElementById('welcome');
    welcome.onmouseover = e => e.target.src = welcomePath;
    welcome.onmouseout = e => e.target.src = welcomeStaticPath;

    const nextImageStaticPath = './assets/img/next_static.gif';
    const nextImagePath = './assets/img/next.gif';
    const nextImage = document.getElementById('send-image');
    nextImage.onmouseover = e => e.target.src = nextImagePath;
    nextImage.onmouseout = e => e.target.src = nextImageStaticPath;
}

async function handleUserInput(event) {
    event.preventDefault();
    nextButton.disabled = true;

    if (!validateUserInput(input)) {
        nextButton.disabled = false;
        return
    }

    toggleLoader();

    let currElement;
    let id;
    await fetch("http://localhost:8081/api/v1/calculate", {
        method: "POST",
        headers: {
            'Content-Type': 'text/plain' // эээхххууухххх...
        },
        body: JSON.stringify({ expression: input.value }),
    })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Network response was not ok? status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            id = data.id;
            currElement = createElement(input.value, id);
            if (expressions.querySelector('.entry-list__placeholder')) {
                expressions.removeChild(document.querySelector('.entry-list__placeholder'));
            }
            expressions.insertBefore(currElement, expressions.firstChild);
            setTimeout(makeFetchRequest, 1000, id, currElement);
            input.value = '';

        })
        .catch(error => {
            console.error(error);
            alert(`Произошла ошибка при отпрравке: ${error}`);
        });

    nextButton.disabled = false;
    toggleLoader();

}

function makeFetchRequest(id, currElement) {
    fetch(`http://localhost:8081/api/v1/expressions/${id}`)
        .then(response => {
            if (!response.ok) {
                throw new Error(`Network response was not ok? status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            if (data.expression.status != "pending") {
                let status = currElement.querySelector('.entry__status');
                status.textContent = 'status: ' + data.expression.status;

                if (data.expression.status == "resolved") {
                    let result = currElement.querySelector('.entry__result');
                    result.textContent = 'result: ' + data.expression.result;
                }
            } else {
                setTimeout(makeFetchRequest, 3000, id, currElement);
            }
        })
        .catch(error => {
            console.error(error);
        });
}

function validateUserInput(input) {
    const regex = /^[ 0-9.()+\-*\/]+$/;
    if (!regex.test(input.value)) {
        nextButton.classList.add('shake');
        input.style.color = 'red';
        setTimeout(() => { nextButton.classList.remove('shake'); input.style.color = null; }, 400);
        return false;
    }

    return true;
}

function toggleLoader() {
    const loader = document.getElementById('loader');
    loader.classList.toggle('hidden');
    nextButton.classList.toggle('hidden');
}

function createElement(expression, id) {

    const container = document.createElement('article');
    container.classList.add('entry');
    container.id = `entry-${id}`;

    const entryInput = document.createElement('div');
    entryInput.classList.add('entry__input');
    entryInput.textContent = expression;
    container.appendChild(entryInput);

    const entryId = document.createElement('div');
    entryId.classList.add('entry__id');
    entryId.textContent = 'id: ' + id;
    container.appendChild(entryId);

    const entryStatus = document.createElement('div');
    entryStatus.classList.add('entry__status');
    entryStatus.textContent = 'status: pending';
    container.appendChild(entryStatus);

    const entryResult = document.createElement('div');
    entryResult.classList.add('entry__result');
    entryResult.textContent = 'result: --/--';
    container.appendChild(entryResult);

    return container;
}
