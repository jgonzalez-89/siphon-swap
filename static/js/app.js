// Swap Manager Class
class SwapManager {
    constructor() {
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.setupNavbarScroll();
        this.loadCurrencies();
    }

    setupEventListeners() {
        // Amount input listener
        const fromAmount = document.getElementById('fromAmount');
        if (fromAmount) {
            fromAmount.addEventListener('input', this.handleAmountChange.bind(this));
        }

        // Currency select listeners
        const fromSelect = document.getElementById('fromSelect');
        const toSelect = document.getElementById('toSelect');
        
        if (fromSelect) {
            fromSelect.addEventListener('change', this.updateFromIcon.bind(this));
        }
        
        if (toSelect) {
            toSelect.addEventListener('change', this.updateToIcon.bind(this));
        }

        // Smooth scroll for anchor links
        document.querySelectorAll('a[href^="#"]').forEach(anchor => {
            anchor.addEventListener('click', function (e) {
                e.preventDefault();
                const target = document.querySelector(this.getAttribute('href'));
                if (target) {
                    target.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                }
            });
        });
    }

    setupNavbarScroll() {
        window.addEventListener('scroll', function() {
            const navbar = document.getElementById('navbar');
            if (window.scrollY > 50) {
                navbar.classList.add('scrolled');
            } else {
                navbar.classList.remove('scrolled');
            }
        });
    }

    loadCurrencies() {
        console.log('Loading currencies...');
        
        // This is now handled by HTMX in the template
        // but we keep the fallback here
        setTimeout(() => {
            const fromSelect = document.getElementById('fromSelect');
            const toSelect = document.getElementById('toSelect');
            
            if (fromSelect && fromSelect.innerHTML.includes('Loading')) {
                // Fallback if HTMX fails
                const fallback = `
                    <option value="btc">BTC</option>
                    <option value="eth">ETH</option>
                    <option value="usdt">USDT</option>
                `;
                fromSelect.innerHTML = fallback;
                toSelect.innerHTML = fallback;
                
                fromSelect.value = 'btc';
                toSelect.value = 'eth';
                this.updateFromIcon();
                this.updateToIcon();
            }
        }, 3000);
    }

    handleAmountChange(event) {
        const amount = parseFloat(event.target.value);
        const button = document.getElementById('swapButton');
        
        if (amount && amount > 0) {
            button.disabled = false;
            button.querySelector('span').textContent = 'Get Best Quote';
        } else {
            button.disabled = true;
            button.querySelector('span').textContent = 'Enter Amount To Preview Swap';
            document.getElementById('toAmount').value = '';
        }
    }

    updateFromIcon() {
        const fromSelect = document.getElementById('fromSelect');
        const fromIcon = document.getElementById('fromTokenIcon');
        if (fromSelect && fromIcon) {
            fromIcon.textContent = fromSelect.value.toUpperCase().slice(0, 3);
        }
    }

    updateToIcon() {
        const toSelect = document.getElementById('toSelect');
        const toIcon = document.getElementById('toTokenIcon');
        if (toSelect && toIcon) {
            toIcon.textContent = toSelect.value.toUpperCase().slice(0, 3);
        }
    }
}

// Global functions that can be called from HTML
function swapPairs() {
    const fromSelect = document.getElementById('fromSelect');
    const toSelect = document.getElementById('toSelect');
    const fromAmount = document.getElementById('fromAmount');
    const toAmount = document.getElementById('toAmount');
    
    const tempValue = fromSelect.value;
    fromSelect.value = toSelect.value;
    toSelect.value = tempValue;
    
    document.getElementById('fromTokenIcon').textContent = fromSelect.value.toUpperCase().slice(0, 3);
    document.getElementById('toTokenIcon').textContent = toSelect.value.toUpperCase().slice(0, 3);
    
    toAmount.value = '';
    
    if (fromAmount.value) {
        htmx.trigger(fromAmount, 'keyup');
    }
}

function executeSwap() {
    const from = document.getElementById('fromSelect').value;
    const to = document.getElementById('toSelect').value;
    const amount = document.getElementById('fromAmount').value;
    const address = document.getElementById('toAddress').value;
    const exchange = document.getElementById('swapButton').getAttribute('data-exchange') || '';
    
    if (!address) {
        alert('Please enter your receiving wallet address');
        document.querySelector('details').open = true;
        document.getElementById('toAddress').focus();
        return;
    }
    
    const formData = new FormData();
    formData.append('from', from);
    formData.append('to', to);
    formData.append('amount', amount);
    formData.append('to_address', address);
    formData.append('exchange', exchange);
    
    fetch('/htmx/swap', {
        method: 'POST',
        body: formData,
        headers: {
            'HX-Request': 'true'
        }
    })
    .then(response => response.text())
    .then(html => {
        document.querySelector('.swap-card').innerHTML = html;
    })
    .catch(error => {
        alert('Error: ' + error);
    });
}

function toggleMobileMenu() {
    const menu = document.getElementById('mobileMenu');
    menu.classList.toggle('active');
}

function selectExchange(exchange) {
    const button = document.getElementById('swapButton');
    const buttonText = button.querySelector('span') || button;
    buttonText.textContent = 'Swap via ' + exchange;
    button.setAttribute('data-exchange', exchange);
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    new SwapManager();
});