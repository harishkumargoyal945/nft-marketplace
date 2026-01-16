const API_URL = "http://localhost:8080/v1";

// Utils
const $ = (id) => document.getElementById(id);

// State
let listings = [];
let currentView = 'marketplace'; // marketplace, my-nfts
let currentUser = { id: 1, wallet_address: "0x..." }; // Mock user for demo

// Init
async function init() {
    try {
        // Try to get or create a mock user for demo
        const userRes = await fetch(`${API_URL}/users`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ wallet_address: "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80", name: "Demo User" })
        });
        if (userRes.ok) {
            currentUser = await userRes.json();
        }

        $('user-wallet').innerText = currentUser.wallet_address ? `🟢 ${currentUser.wallet_address.slice(0, 6)}...${currentUser.wallet_address.slice(-4)}` : 'Not Connected';

        loadListings();
        setupEventListeners();
    } catch (err) {
        console.error(err);
        $('user-wallet').innerText = '🔴 API Error';
    }
}

// Setup Event Listeners
function setupEventListeners() {
    // Tab switching
    document.querySelectorAll('.view-tab').forEach(tab => {
        tab.addEventListener('click', (e) => {
            document.querySelectorAll('.view-tab').forEach(t => t.classList.remove('active'));
            e.target.classList.add('active');
            currentView = e.target.dataset.view;

            if (currentView === 'marketplace') {
                loadListings();
            } else if (currentView === 'my-nfts') {
                loadMyNFTs();
            }
        });
    });

    // Launch NFT form
    const launchForm = $('launch-form');
    if (launchForm) {
        launchForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const name = $('coin-name').value;
            const symbol = $('coin-ticker').value;
            const desc = $('coin-desc').value;
            const image = $('coin-image').value;

            const btn = e.target.querySelector('button');
            const originalText = btn.innerText;
            btn.innerText = '🚀 Minting...';
            btn.disabled = true;

            try {
                const res = await fetch(`${API_URL}/nfts/mint`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        owner_id: currentUser.id,
                        name: name,
                        symbol: symbol,
                        description: desc,
                        image_url: image
                    })
                });

                if (res.ok) {
                    const nft = await res.json();
                    showToast(`✅ NFT #${nft.token_id} minted successfully!`, 'success');
                    e.target.reset();
                    if (currentView === 'my-nfts') {
                        loadMyNFTs();
                    }
                } else {
                    const err = await res.json();
                    throw new Error(err.error || "Minting failed");
                }
            } catch (err) {
                showToast(`❌ Minting failed: ${err.message}`, 'error');
            } finally {
                btn.innerText = originalText;
                btn.disabled = false;
            }
        });
    }
}

// Load Listings (Marketplace View)
async function loadListings() {
    try {
        const res = await fetch(`${API_URL}/listings`);
        listings = await res.json();
        renderMarketplace(listings);
    } catch (err) {
        console.error("Failed to load listings", err);
        $('coin-grid').innerHTML = '<div class="loading">Failed to load marketplace</div>';
    }
}

// Load My NFTs
async function loadMyNFTs() {
    try {
        const res = await fetch(`${API_URL}/nfts?owner_id=${currentUser.id}`);
        const myNFTs = await res.json();
        renderMyNFTs(myNFTs);
    } catch (err) {
        console.error("Failed to load my NFTs", err);
        $('coin-grid').innerHTML = '<div class="loading">Failed to load your NFTs</div>';
    }
}

function renderMarketplace(list) {
    const grid = $('coin-grid');
    grid.innerHTML = '';

    if (!list || list.length === 0) {
        grid.innerHTML = '<div class="loading">No listings active yet. Be the first! 🚀</div>';
        return;
    }

    list.forEach(listing => {
        const card = document.createElement('div');
        card.className = 'coin-card';
        card.onclick = () => openTradeModal(listing);
        const imageUri = listing.nft.metadata_url || '';
        card.innerHTML = `
            <img src="${imageUri}" class="coin-img" onerror="this.src='https://placehold.co/200x200/667eea/white?text=NFT'">
            <div class="coin-info">
                <h4>NFT #${listing.nft.token_id}</h4>
                <div class="ticker">Price: ${parseFloat(listing.price_wei) / 1e18} ${listing.currency}</div>
                <div class="coin-mc">Seller: ${listing.seller.name || listing.seller.wallet_address.slice(0, 6)}...</div>
            </div>
        `;
        grid.appendChild(card);
    });
}

function renderMyNFTs(list) {
    const grid = $('coin-grid');
    grid.innerHTML = '';

    if (!list || list.length === 0) {
        grid.innerHTML = '<div class="loading">You don\'t own any NFTs yet. Mint one to get started! 🎨</div>';
        return;
    }

    list.forEach(nft => {
        const card = document.createElement('div');
        card.className = 'coin-card my-nft-card';
        card.innerHTML = `
            <img src="${nft.metadata_url}" class="coin-img" onerror="this.src='https://placehold.co/200x200/667eea/white?text=NFT'">
            <div class="coin-info">
                <h4>NFT #${nft.token_id}</h4>
                <div class="ticker">Contract: ${nft.contract_address.slice(0, 10)}...</div>
                <span class="badge owned-badge">Owned</span>
            </div>
            <div class="nft-actions">
                <button class="btn-trade" data-nft-id="${nft.id}">List for Sale</button>
            </div>
        `;

        const listBtn = card.querySelector('.btn-trade');
        listBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            showListModal(nft);
        });

        grid.appendChild(card);
    });
}

function showListModal(nft) {
    const price = prompt("Enter listing price in ETH:", "0.1");
    if (!price) return;
    createListing(nft.id, price);
}

async function createListing(nftId, priceEth) {
    const wei = (parseFloat(priceEth) * 1e18).toString();
    try {
        const res = await fetch(`${API_URL}/listings`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                nft_id: nftId,
                seller_user_id: currentUser.id,
                price_wei: wei,
                currency: "ETH"
            })
        });
        if (res.ok) {
            showToast("✅ NFT listed successfully!", "success");
            loadListings();
        } else {
            const err = await res.json();
            throw new Error(err.error);
        }
    } catch (err) {
        showToast(`❌ Listing failed: ${err.message}`, "error");
    }
}

// Modal Logic
const modal = $('trade-modal');
const closeBtn = document.querySelector('.close-modal');

if (closeBtn) {
    closeBtn.onclick = () => {
        modal.classList.add('hidden');
    }
}

window.onclick = (event) => {
    if (event.target == modal) {
        modal.classList.add('hidden');
    }
}

function openTradeModal(listing) {
    $('modal-img').src = listing.nft.metadata_url;
    $('modal-name').innerText = `NFT #${listing.nft.token_id}`;
    $('modal-ticker').innerText = listing.nft.contract_address;
    $('modal-mc').innerText = `${parseFloat(listing.price_wei) / 1e18} ${listing.currency}`;
    $('modal-creator').innerText = listing.seller.wallet_address;
    $('modal-desc').innerText = `Chain: ${listing.nft.chain}`;
    $('modal-token-id').value = listing.id;
    $('trade-result').classList.add('hidden');

    modal.classList.remove('hidden');
}

// Trade/Buy Logic
const tradeForm = $('trade-form');
if (tradeForm) {
    tradeForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const listingId = parseInt($('modal-token-id').value);

        const btn = $('trade-btn');
        const originalText = btn.innerText;
        btn.innerText = 'Processing...';
        btn.disabled = true;

        try {
            const orderRes = await fetch(`${API_URL}/orders`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ listing_id: listingId, buyer_user_id: currentUser.id })
            });
            const order = await orderRes.json();

            if (order.id) {
                const confirmRes = await fetch(`${API_URL}/orders/${order.id}/confirm`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ tx_hash: "0x" + Math.random().toString(16).slice(2) })
                });
                const result = await confirmRes.json();

                if (result.status === "confirmed") {
                    showToast(`✅ Successfully bought NFT!`, 'success');
                    loadListings();
                    setTimeout(() => modal.classList.add('hidden'), 2000);
                } else {
                    throw new Error(result.error || "Confirmation failed");
                }
            } else {
                throw new Error(order.error || "Order creation failed");
            }
        } catch (err) {
            showToast(`❌ Error: ${err.message}`, 'error');
        } finally {
            btn.innerText = originalText;
            btn.disabled = false;
        }
    });
}

// Toast Notifications
function showToast(message, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.innerText = message;
    document.body.appendChild(toast);

    setTimeout(() => toast.classList.add('show'), 100);
    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

init();
