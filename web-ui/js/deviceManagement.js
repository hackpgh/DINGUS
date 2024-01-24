document.getElementById('deviceManagementForm').addEventListener('submit', function(e) {
    e.preventDefault();

    if (!validateForm()) {
        return;
    }

    let formData = new FormData(this);

    fetch(this.action, {
        method: this.method,
        body: formData,
    })
    .then(response => {
        if (response.ok) {
            showToast("Device assignments updated successfully.");
        } else {
            showToast("Failed to update device assignments.");
        }
    })
    .catch(() => {
        showToast("An error occurred. Please try again.");
    });
});

function validateForm() {
    let isValid = true;
    let trainingLabelSet = new Set();
    let allSelects = document.querySelectorAll('#deviceList select');

    allSelects.forEach(select => {
        if (select.value === '') {
            showToast("Please assign a training label to all devices.");
            isValid = false;
            return;
        }
        if (trainingLabelSet.has(select.value)) {
            showToast("Each training label can only be assigned to one device.");
            isValid = false;
            return;
        }
        trainingLabelSet.add(select.value);
    });

    return isValid;
}

function showToast(message) {
    let toastElement = document.querySelector('.toast');
    let toastBody = toastElement.querySelector('.toast-body');
    toastBody.textContent = message;

    $(toastElement).toast({ delay: 3000 });
    $(toastElement).toast('show');
}