// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

contract Marketplace is ReentrancyGuard {
    struct Listing {
        uint256 price;
        address seller;
        bool active;
    }

    // NFT Address -> Token ID -> Listing
    mapping(address => mapping(uint256 => Listing)) public listings;

    event Listed(address indexed nft, uint256 indexed tokenId, uint256 price, address indexed seller);
    event Bought(address indexed nft, uint256 indexed tokenId, uint256 price, address indexed buyer);
    event Delisted(address indexed nft, uint256 indexed tokenId, address indexed seller);

    function list(address nft, uint256 tokenId, uint256 price) external nonReentrant {
        IERC721 token = IERC721(nft);
        require(token.ownerOf(tokenId) == msg.sender, "Not owner");
        require(token.getApproved(tokenId) == address(this) || token.isApprovedForAll(msg.sender, address(this)), "Not approved");
        require(price > 0, "Price must be > 0");

        listings[nft][tokenId] = Listing(price, msg.sender, true);
        emit Listed(nft, tokenId, price, msg.sender);
    }

    function delist(address nft, uint256 tokenId) external nonReentrant {
        Listing storage item = listings[nft][tokenId];
        require(item.active, "Not listed");
        require(item.seller == msg.sender, "Not seller");

        item.active = false;
        emit Delisted(nft, tokenId, msg.sender);
    }

    function buy(address nft, uint256 tokenId) external payable nonReentrant {
        Listing memory item = listings[nft][tokenId];
        require(item.active, "Not for sale");
        require(msg.value >= item.price, "Insufficient funds");

        listings[nft][tokenId].active = false; // Delist

        IERC721(nft).safeTransferFrom(item.seller, msg.sender, tokenId);

        // Pay seller
        (bool success, ) = payable(item.seller).call{value: item.price}("");
        require(success, "Transfer failed");
        
        emit Bought(nft, tokenId, item.price, msg.sender);
    }

    function getListing(address nft, uint256 tokenId) external view returns (Listing memory) {
        return listings[nft][tokenId];
    }
}
