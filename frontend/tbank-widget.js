(function (global) {
    if (!window.payForm) {
        throw new Error("tinkoff_v2.js not loaded");
    }

    function pay({ paymentId, mode = "popup" }) {
        if (mode === "redirect") {
            window.location.href =
                "https://securepayments.tbank.ru/eacq/v2/Pay?PaymentId=" +
                paymentId;
            return;
        }

        payForm.open({
            paymentId: paymentId,
            language: "ru",
        });
    }

    global.TBank = { pay };
})(window);
