document.getElementById('deviceManagementForm').addEventListener('submit', function(e) {
    e.preventDefault();

    if (!validateForm()) {
        return;
    }

    // Initialize an array to hold the assignment objects
    let assignments = [];
    document.querySelectorAll('#deviceList tr').forEach(row => {
        let ipAddress = row.querySelector('td:nth-child(1)').textContent.trim();
        let macAddress = row.querySelector('td:nth-child(2)').textContent.trim(); // No need to replace colons for JSON
        let selectedTraining = row.querySelector('select').value;

        // Push an object for each row into the assignments array
        assignments.push({
            ipAddress: ipAddress,
            macAddress: macAddress,
            trainingLabel: selectedTraining
        });
    });

    fetch(this.action, {
        method: this.method,
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(assignments), // Convert the array to JSON
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showToast("Device assignments updated successfully.");
        } else {
            showToast(data.message || "Failed to update device assignments.");
        }
    })
    .catch(() => {
        showToast("An error occurred. Please try again.");
    });
});


function validateForm() {
    let isValid = true;
    let allSelects = document.querySelectorAll('#deviceList select');

    // Reset validation state
    allSelects.forEach(select => select.classList.remove('is-invalid'));

    allSelects.forEach(select => {
        if (select.value === '') {
            // Highlight the invalid select element
            select.classList.add('is-invalid');
            showToast("Please assign a training label to all devices.");
            isValid = false;
        }
    });

    return isValid;
}

function showToast(message) {
    let toastElement = document.querySelector('.toast');
    let toastBody = toastElement.querySelector('.toast-body') || toastElement; // Fallback to toastElement if .toast-body not found
    toastBody.textContent = message;

    // Use Bootstrap's Toast component if available, otherwise fallback
    if (typeof bootstrap !== 'undefined' && bootstrap.Toast) {
        let toast = new bootstrap.Toast(toastElement);
        toast.show();
    } else {
        // Fallback or custom toast display logic
        toastElement.style.display = 'block';
        setTimeout(() => toastElement.style.display = 'none', 3000);
    }
}
