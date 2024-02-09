document.getElementById('configForm').addEventListener('submit', function(e) {
    e.preventDefault();

    var accountIdValue = document.getElementById('wildApricotAccountId').value.trim();
    var accountId = parseInt(accountIdValue, 10);

    if (isNaN(accountId) || accountId <= 0 || accountIdValue.length > 10) {
        alert('Invalid Wild Apricot Account ID. Please enter a valid number.');
        return;
    }

    var configData = {
        cert_file: document.getElementById('certFile').value,
        key_file: document.getElementById('keyFile').value,
        wild_apricot_account_id: accountId,
        contact_filter_query: document.getElementById('contactFilterQuery').value,
        tag_id_field_name: document.getElementById('tagIdFieldName').value,
        training_field_name: document.getElementById('trainingFieldName').value,
        loki_hook_url: document.getElementById('lokiHookUrl').value,
    };

    fetch('/api/updateConfig', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(configData),
    })
    .then(response => {
        if(response.ok) {
            alert('Configuration updated successfully');
        } else {
            alert('Failed to update configuration');
        }
    });
});