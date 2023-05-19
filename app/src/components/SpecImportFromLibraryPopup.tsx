import React, { useState } from "react";
import { Button, Modal } from "react-bootstrap";

import { SpecLibraryList } from "./SpecLibraryList";

export function SpecImportFromLibraryButton() {
    const [show, setShow] = useState(false);
    const handleClose = () => setShow(false);
    const handleShow = () => setShow(true);

    return (
        <>
            <Button className="m-1" variant="secondary" onClick={handleShow}>
                Import
            </Button>

            <Modal centered size="lg" show={show} onHide={handleClose}>
                <Modal.Header closeButton>
                    <Modal.Title>Import Spec</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <SpecLibraryList />
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={handleClose}>
                        Close
                    </Button>
                </Modal.Footer>
            </Modal>
        </>
    );
}
